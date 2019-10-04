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
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import {Loader} from "./Loader";
import {BountyState, BountyStore, mapStateToStr} from "../stores/BountyStore";
import {Link} from 'react-router-dom';
import * as dateformat from 'dateformat';
import {UIStore} from "../stores/UIStore";
import {Redirect} from "react-router";

interface Props {
    userStore?: UserStore;
    repoStore?: RepositoryStore;
    uiStore?: UIStore;
    bountyStore?: BountyStore;
    match?: {
        params: {
            id: string,
        }
    }
}

@inject("userStore")
@inject("bountyStore")
@inject("uiStore")
@inject("repoStore")
@observer
export default class Bounty extends React.Component<Props, {}> {

    componentWillMount() {
        let {id} = this.props.match.params;
        this.props.bountyStore.fetchBounty(id);
        this.props.repoStore.fetchRepoForBounty(id);
    }

    componentWillUnmount() {
        this.closeDeleteBountyModal();
        this.props.bountyStore.resetDeleted();
    }

    deleteBounty = () => {
        let {id} = this.props.bountyStore.bounty;
        this.props.bountyStore.deleteBounty(id);
    }

    openDeleteBountyModal = () => {
        this.props.uiStore.setDeleteBountyModalOpen(true);
    }

    closeDeleteBountyModal = () => {
        this.props.uiStore.setDeleteBountyModalOpen(false);
    }

    render() {
        let {bounty} = this.props.bountyStore;
        let {repo} = this.props.repoStore;
        let {deleteBountyModalOpen} = this.props.uiStore;

        if (this.props.bountyStore.deleted) {
            return <Redirect to={`/repo/${repo.owner}/${repo.name}`}/>;
        }

        if (this.props.bountyStore.loading || this.props.repoStore.loading) {
            return <Loader/>;
        }

        return (
            <React.Fragment>
                <Dialog
                    open={deleteBountyModalOpen}
                    maxWidth={"md"}
                >
                    <DialogTitle>{"Delete Bounty"}</DialogTitle>
                    <DialogContent>
                        <DialogContentText>
                            Are you sure you want to delete the bounty? The funds pooled on the generated
                            receiving address can not be recovered.
                        </DialogContentText>
                        <DialogContentText>
                            Deleting the bounty will automatically post a message under the issue on GitHub
                            that the bounty is no longer active.
                        </DialogContentText>
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={this.deleteBounty} color="primary">
                            Yes
                        </Button>
                        <Button onClick={this.closeDeleteBountyModal} color="primary">
                            No
                        </Button>
                    </DialogActions>
                </Dialog>
                <Grid container className={css.dashboard} justify="flex-start" spacing={16}>
                    <Grid item xs={12}>
                        <Typography component="h2">
                            <Link to={`/`}>Repositories</Link><ArrowRight className={css.verticalAlign}/>
                            <Link to={`/repo/${repo.owner}/${repo.name}`}>{`Repository `}</Link>
                            <a className={css.underlined} href={`https://github.com/${repo.owner}/${repo.name}`}
                               target={'_blank'}>
                                {repo.owner} / {repo.name}
                            </a>
                            <ArrowRight className={css.verticalAlign}/>
                            {`Bounty for `}
                            <a className={css.underlined} href={`https://github.com/${repo.owner}/${repo.name}/issues/${bounty.issue_number}`}
                               target={'_blank'}>
                                {bounty.title}
                            </a>
                        </Typography>
                        <Divider className={css.dividerSmall}/>
                    </Grid>
                    <Grid item xs={12} className={css.marginBottom}>
                        <Button variant="outlined" color="secondary" onClick={this.openDeleteBountyModal}>
                            Remove from bounty system
                        </Button>
                    </Grid>
                    <Grid item xs={12}>
                        <Typography color="textSecondary" className={css.marginBottom}>
                            Linked to issue with ID: {bounty.id}
                        </Typography>
                        <Typography component="h2">
                            State
                        </Typography>
                        {
                            bounty.state == BountyState.Transferred ?
                                <div>
                                    <Typography component="p">
                                        {`${mapStateToStr(bounty.state)} `}
                                        <a className={css.underlined}
                                           href={`https://thetangle.org/bundle/${bounty.bundle_hash}`}
                                           target={'_blank'}>
                                            bundle
                                        </a>
                                        {` to `}
                                        <a className={css.underlined}
                                           href={`https://thetangle.org/address/${bounty.receiver_address}`}
                                           target={'_blank'}>
                                            target address.
                                        </a>
                                    </Typography>
                                </div>
                                :
                                <Typography component="p">
                                    {mapStateToStr(bounty.state)}
                                </Typography>
                        }
                        <Divider className={css.dividerMiddle}/>
                        <Typography component="h2">
                            Bounty Address
                        </Typography>
                        <Grid item xs={6}>
                            <a href={`https://thetangle.org/address/${bounty.pool_address}`} target={'_blank'}>
                                <div className={css.addressBox}>{bounty.pool_address}</div>
                            </a>
                        </Grid>
                        <Divider className={css.dividerMiddle}/>
                        <Typography component="h2">
                            Balance
                        </Typography>
                        <Typography component="p">
                            {bounty.balance} iotas
                        </Typography>
                        <Divider className={css.dividerMiddle}/>
                        {
                            bounty.body !== '' &&
                            <div>
                                <Typography component="h3">
                                    Issue Text
                                </Typography>
                                <Typography component="p">
                                    {bounty.body}
                                </Typography>
                                <Divider className={css.dividerMiddle}/>
                            </div>
                        }
                        <Typography color="textSecondary">
                            Created on: {dateformat(bounty.created_on, "dd.mm.yyyy HH:MM:ss")}
                            <br/>
                            {
                                bounty.updated_on !== null &&
                                <span>Last updated: {dateformat(bounty.updated_on, "dd.mm.yyyy HH:MM:ss")}</span>
                            }
                        </Typography>
                    </Grid>
                </Grid>
            </React.Fragment>
        );
    }
}