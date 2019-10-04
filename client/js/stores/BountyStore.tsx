import {action, observable} from 'mobx';
import {CreateError, FormState, mapTextToError} from "../misc/Misc";
import {Model} from "./AppStore";

export enum BountyState {
    Open,
    Released,
    Transferred
}

export function mapStateToStr(state: BountyState): string {
    switch (state) {
        case BountyState.Open:
            return "Open";
        case BountyState.Released:
            return "Released";
        case BountyState.Transferred:
            return "Transferred";
        default:
            return "Unknown"
    }
}

export class Bounty extends Model {
    id: number;
    issue_number: number;
    repository_id: number;
    receiver_id: number;
    pool_address: string;
    receiver_address: string;
    bundle_hash: string;
    balance: number;
    url: string;
    title: string;
    body: string;
    state: BountyState;
}

export let BountyCreateError = {
    ...CreateError,
    IssueClosed: "issue is closed",
    IssueDoesntExist: "issue doesn't exist",
    RepositoryNotInPlatform: "repository not added to platform",
}

let errorTextMap = {
    [CreateError.Unknown]: "An unknown error occurred.",
    [CreateError.AlreadyExists]: "The issue is already added to the platform.",
    [BountyCreateError.IssueClosed]: "The issue is already closed. You can only add open issues.",
    [BountyCreateError.IssueDoesntExist]: "The issue wasn't found on the repository, did you perhaps mean another issue?",
    [BountyCreateError.RepositoryNotInPlatform]: "The repository to which this issue belongs to is not part of the platform",
    [CreateError.NotFound]: "The bounty doesn't exist.",
};

export class BountyStore {

    @observable err: any = null;
    @observable loading: boolean;
    @observable deleted: boolean;

    // bounties
    @observable bounties = new Map();

    // single bounty
    @observable bounty: Bounty = null;

    // new bounty
    @observable new_bounty_issue_id: number = null;
    @observable new_bounty_form_state = FormState.Init;

    @action
    resetDeleted = () => this.deleted = false;

    @action
    setDeleted = (deleted: boolean) => this.deleted = deleted;

    fetchBounty = async (id: string) => {
        this.setLoading(true);
        try {
            let res = await fetch(`/api/bounties/${id}`);
            if (res.status !== 200) {
                let errorTxt = await res.text();
                this.setError(errorTxt);
                return;
            }
            let bounty: Bounty = await res.json();
            this.setBounty(bounty);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    fetchBountiesOfRepo = async (owner: string, name: string) => {
        this.setLoading(true);
        try {
            let res = await fetch(`/api/bounties/${owner}/${name}`);
            if (res.status !== 200) {
                let errorText = await res.text();
                this.setError(errorText);
                return;
            }
            let bounties: Array<Bounty> = await res.json();
            this.setBounties(bounties);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    deleteBounty = async (id: number) => {
        this.setLoading(true);
        try {
            await fetch(`/api/bounties/${id}`, {method: 'DELETE'});
            this.setDeleted(true);
        } catch (err) {
            this.setError(err);
        } finally {
            this.setLoading(false);
        }
    }

    addBounty = async (owner: string, name: string) => {
        this.setLoading(true);
        try {
            let res = await fetch(`/api/bounties?issue_id=${this.new_bounty_issue_id}&owner=${owner}&name=${name}`, {method: 'POST'});
            if (res.status !== 200) {
                let errorText = await res.text();
                this.setError(mapTextToError(errorText, BountyCreateError, errorTextMap));
                return;
            }
            let bounty: Bounty = await res.json();
            this.setBounty(bounty);
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
    setBounty = (bounty: Bounty) => this.bounty = bounty;

    @action
    updateNewRepoFormState = (newState: FormState) => this.new_bounty_form_state = newState;

    @action
    setBounties = (bounties: Array<Bounty>) => {
        let newMap = new Map();
        bounties.forEach(bounty => newMap.set(bounty.id, bounty));
        this.bounties = newMap;
    }

    @action
    resetFormData = () => {
        this.new_bounty_issue_id = null;
        this.new_bounty_form_state = FormState.Init;
        this.loading = false;
        this.err = null;
    }

    @action
    updateNewBountyIssueID = (id: string) => {
        if (id === '') {
            this.new_bounty_form_state = FormState.Init;
            this.new_bounty_issue_id = null;
            return;
        }
        this.new_bounty_issue_id = parseInt(id);
        this.new_bounty_form_state = FormState.Ok;
    }

}

export var BountyStoreInstance = new BountyStore();