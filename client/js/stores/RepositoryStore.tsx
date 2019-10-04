import {action, observable} from 'mobx';
import {isValidGitHubURL} from "../misc/Utils";
import {CreateError, FormState, mapTextToError} from "../misc/Misc";
import {Model} from "./AppStore";

export class Repository extends Model {
    id: number;
    owner: string;
    name: string;
    description: string;
}

export let RepoCreateError = {
    ...CreateError,
    IssuesDeactivated: "repository has issues deactivated",
}

let errorTextMap = {
    [CreateError.Unknown]: "An unknown error occurred.",
    [RepoCreateError.IssuesDeactivated]: "The repository must have issues activated.",
    [CreateError.AlreadyExists]: "The repository already exists.",
    [CreateError.NotFound]: "The repository doesn't exist.",
};

export class RepositoryStore {

    @observable err: any;
    @observable loading: boolean;
    @observable deleted: boolean;

    // repositories
    @observable repositories = new Map();

    // single repository
    @observable repo: Repository = null;

    // new repository
    @observable new_repo_url: string = "";
    @observable new_repo_form_state = FormState.Init;

    @action
    resetDeleted = () => this.deleted = false;

    @action
    setDeleted = (deleted: boolean) => this.deleted = deleted;


    fetchRepo = async (owner: string, name: string) => {
        this.setLoading(true);
        try {
            let res = await fetch(`/api/repos/${owner}/${name}`);
            if (res.status !== 200) {
                let errorTxt = await res.text();
                this.setError(errorTxt);
                return;
            }
            let repo: Repository = await res.json();
            this.setRepo(repo);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    fetchRepoForBounty = async (id: string) => {
        this.setLoading(true);
        try {
            let res = await fetch(`/api/repos/of/${id}`);
            if (res.status !== 200) {
                let errorTxt = await res.text();
                this.setError(errorTxt);
                return;
            }
            let repo: Repository = await res.json();
            this.setRepo(repo);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    deleteRepo = async (id: number) => {
        this.setLoading(true);
        try {
            await fetch(`/api/repos/${id}`, {method: 'DELETE'});
            this.setDeleted(true);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    fetchRepos = async () => {
        this.setLoading(true);
        try {
            let res = await fetch('/api/repos');
            if (res.status !== 200) {
                let errorText = await res.text();
                this.setError(errorText);
                return;
            }
            let repos: Array<Repository> = await res.json();
            this.setRepos(repos);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    addRepo = async () => {
        this.setLoading(true);
        try {
            let res = await fetch(`/api/repos?url=${this.new_repo_url}`, {method: 'POST'});
            if (res.status !== 200) {
                let errorText = await res.text();
                this.setError(mapTextToError(errorText, RepoCreateError, errorTextMap));
                return;
            }
            let repo: Repository = await res.json();
            this.setRepo(repo);
            this.updateNewRepoFormState(FormState.Finished);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    @action
    setError = (err: any) => this.err = err;

    @action
    setLoading = (loading: boolean) => this.loading = loading;

    @action
    setRepo = (repo: Repository) => this.repo = repo;

    @action
    resetRepo = () => this.repo = null;

    @action
    setRepos = (repos: Array<Repository>) => {
        let newMap = new Map();
        repos.forEach(repo => newMap.set(repo.id, repo));
        this.repositories = newMap;
    }

    @action
    updateNewRepoFormState = (newState: FormState) => this.new_repo_form_state = newState;

    @action
    resetFormData = () => {
        this.new_repo_url = "";
        this.new_repo_form_state = FormState.Init;
        this.loading = false;
        this.err = null;
    }

    @action
    updateNewRepoURL = (url: string) => {
        this.new_repo_url = url;
        if (isValidGitHubURL(this.new_repo_url)) {
            this.new_repo_form_state = FormState.Ok;
            return;
        }
        this.new_repo_form_state = FormState.Invalid;
    }
}

export var RepositoryStoreInstance = new RepositoryStore();