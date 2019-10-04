import * as React from 'react';
import {inject, observer} from 'mobx-react';
import {Redirect} from "react-router";

import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import Divider from '@material-ui/core/Divider';
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import LinearProgress from '@material-ui/core/LinearProgress';

import {RepositoryStore} from "../stores/RepositoryStore";
import {UIStore} from "../stores/UIStore";
import {FormState} from "../misc/Misc";
import {BountyStore} from "../stores/BountyStore";

import * as css from './app.scss';

interface Props {
    repoStore?: RepositoryStore;
    bountyStore?: BountyStore;
    uiStore?: UIStore;
}

@inject("repoStore")
@inject("bountyStore")
@inject("uiStore")
@observer
export class BountyForm extends React.Component<Props, {}> {

    componentWillUnmount() {
        this.props.bountyStore.resetFormData();
        this.props.uiStore.setBountyFormOpen(false);
    }

    updateIssueID = (e) => {
        this.props.bountyStore.updateNewBountyIssueID(e.target.value);
    }

    addBounty = () => {
        let {repo} = this.props.repoStore;
        this.props.bountyStore.addBounty(repo.owner, repo.name);
    }

    render() {
        let {
            new_bounty_issue_id, new_bounty_form_state, loading, bounty, err
        } = this.props.bountyStore;
        if (bounty && new_bounty_form_state === FormState.Finished) {
            return <Redirect to={`/bounty/${bounty.id}`}/>
        }

        return (
            <React.Fragment>
                <Grid container justify="flex-start" spacing={16}>
                    <Grid item xs={12}>
                        <Typography component="h4">
                            New Bounty
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
                            <Grid item xs={4}>
                                <Typography component="p" className={css.errorText}>
                                    {err}
                                </Typography>
                            </Grid>
                        </Grid>
                    }
                    <Grid item xs={4}>
                        <TextField
                            id="new-issue-id"
                            label="Issue ID"
                            value={new_bounty_issue_id || ""}
                            onChange={this.updateIssueID}
                            type="number"
                            placeholder="1234"
                            helperText="The ID of the issue on GitHub"
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
                            onClick={this.addBounty}
                            disabled={
                                new_bounty_form_state !== FormState.Ok
                                ||
                                loading
                            }
                        >
                            Link issue
                        </Button>
                    </Grid>
                </Grid>
            </React.Fragment>
        );
    }
}