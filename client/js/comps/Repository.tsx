import * as React from 'react';
import {inject, observer} from 'mobx-react';
import Grid from '@material-ui/core/Grid';

import {UserStore} from "../stores/UserStore";

import * as css from './app.scss';
import {RepositoryStore} from "../stores/RepositoryStore";
import Divider from '@material-ui/core/Divider';
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import ArrowRight from '@material-ui/icons/KeyboardArrowRight';
import {Loader} from "./Loader";
import Bounties from "./Bounties";
import {Link} from 'react-router-dom';
import Dialog from "@material-ui/core/Dialog";
import DialogTitle from "@material-ui/core/DialogTitle";
import DialogContent from "@material-ui/core/DialogContent";
import DialogContentText from "@material-ui/core/DialogContentText";
import DialogActions from "@material-ui/core/DialogActions";
import {UIStore} from "../stores/UIStore";
import {Redirect} from "react-router";

interface Props {
    userStore?: UserStore;
    repoStore?: RepositoryStore;
    uiStore?: UIStore;
    match?: {
        params: {
            owner: string,
            name: string,
        }
    }
}

@inject("userStore")
@inject("repoStore")
@inject("uiStore")
@observer
export default class Repository extends React.Component<Props, {}> {

    componentWillMount() {
        let {owner, name} = this.props.match.params;
        this.props.repoStore.fetchRepo(owner, name);
    }

    componentWillUnmount() {
        this.closeDeleteRepoModal();
        this.props.repoStore.resetDeleted();
    }

    deleteRepo = () => {
        let {id} = this.props.repoStore.repo;
        this.props.repoStore.deleteRepo(id);
    }

    openDeleteRepoModal = () => {
        this.props.uiStore.setDeleteRepoModalOpen(true);
    }

    closeDeleteRepoModal = () => {
        this.props.uiStore.setDeleteRepoModalOpen(false);
    }

    render() {
        let {repo, loading} = this.props.repoStore;
        let {deleteRepoModalOpen} = this.props.uiStore;

        if (this.props.repoStore.deleted) {
            return <Redirect to={`/`}/>;
        }

        if (loading) {
            return <Loader/>;
        }

        return (
            <React.Fragment>
                <Dialog
                    open={deleteRepoModalOpen}
                    maxWidth={"md"}
                >
                    <DialogTitle>{"Delete Repository"}</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            Are you sure you want to delete the repository? All bounties and their associated
                            funds will be removed from the platform. The funds can not be recovered.
                        </DialogContentText>
                        <DialogContentText>
                            A message will be posted on each issue on GitHub mentioning that the bounty
                            is no longer active.
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.deleteRepo} color="primary">
                            Yes
                        </Button>
                        <Button onClick={this.closeDeleteRepoModal} color="primary">
                            No
                        </Button>
                    </DialogActions>
                </Dialog>
                <Grid container className={css.dashboard} justify="flex-start" spacing={16}>
                    <Grid item xs={12}>
                        <Typography component="h2">
                            <Link to={"/"}>Repositories</Link><ArrowRight className={css.verticalAlign}/>
                            {`Repository `}
                            <a className={css.underlined} href={`https://github.com/${repo.owner}/${repo.name}`}
                               target={'_blank'}>
                                {repo.owner} / {repo.name}
                            </a>
                        </Typography>
                        <Divider className={css.dividerSmall}/>
                    </Grid>
                    {
                        repo.description !== '' &&
                        <Grid item xs={12}>
                            <Typography component="p">
                                {repo.description}
                            </Typography>
                        </Grid>
                    }
                    <Grid item xs={12} className={css.marginBottom}>
                        <Button variant="outlined" color="secondary" onClick={this.openDeleteRepoModal}>
                            Remove from bounty system
                        </Button>
                    </Grid>
                </Grid>

                <Bounties/>
            </React.Fragment>
        );
    }
}