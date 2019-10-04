import * as React from 'react';
import {inject, observer} from 'mobx-react';
import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import {Redirect} from "react-router";

import {withAuth} from "./Authenticated";
import {UserStore} from "../stores/UserStore";

import * as css from './app.scss';
import {RepositoryStore} from "../stores/RepositoryStore";
import Divider from '@material-ui/core/Divider';
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import LinearProgress from '@material-ui/core/LinearProgress';
import {UIStore} from "../stores/UIStore";
import {FormState} from "../misc/Misc";

interface Props {
    userStore?: UserStore;
    repoStore?: RepositoryStore;
    uiStore?: UIStore;
}

@inject("userStore")
@inject("repoStore")
@inject("uiStore")
@observer
export class repositoryForm extends React.Component<Props, {}> {

    componentWillUnmount() {
        this.props.repoStore.resetFormData();
        this.props.uiStore.setRepoFormOpen(false);
    }

    updateURL = (e) => {
        this.props.repoStore.updateNewRepoURL(e.target.value);
    }

    addRepository = () => {
        this.props.repoStore.addRepo();
    }

    render() {
        let {
            new_repo_url, new_repo_form_state,
            loading, repo, err
        } = this.props.repoStore;

        if (repo && new_repo_form_state === FormState.Finished) {
            return <Redirect to={`/repo/${repo.owner}/${repo.name}`}/>
        }

        return (
            <React.Fragment>
                <Grid container className={css.dashboard} justify="flex-start" spacing={16}>
                    <Grid item xs={12}>
                        <Typography component="h4">
                            New Repository
                        </Typography>
                        <Divider className={css.dividerSmall}/>
                    </Grid>
                    {
                        loading &&
                        <Grid item container xs={12}>
                            <Grid item xs={4}>
                                <LinearProgress/>
                            </Grid>
                        </Grid>
                    }
                    {
                        err !== null &&
                        <Grid item container xs={12}>
                            <Grid item xs={12}>
                                <Typography component="p" className={css.errorText}>
                                    {err}
                                </Typography>
                            </Grid>
                        </Grid>
                    }
                    <Grid item xs={4}>
                        <TextField
                            id="new-repo-url"
                            label="Repository URL"
                            value={new_repo_url}
                            onChange={this.updateURL}
                            type="text"
                            placeholder="link"
                            helperText="The URL of the GitHub repository"
                            InputLabelProps={{shrink: true}}
                            disabled={loading}
                            margin="normal"
                            variant="outlined"
                            fullWidth
                        />
                    </Grid>
                    <Grid item xs={12} className={css.marginBottom}>
                        <Button
                            variant="outlined"
                            color="primary"
                            onClick={this.addRepository}
                            disabled={
                                new_repo_form_state !== FormState.Ok
                                ||
                                loading
                            }
                        >
                            Add Repository
                        </Button>
                    </Grid>
                </Grid>
            </React.Fragment>
        );
    }
}

export let RepositoryForm = withAuth(repositoryForm);