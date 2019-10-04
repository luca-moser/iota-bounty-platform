import * as React from 'react';
import {inject, observer} from 'mobx-react';
import clsx from "clsx";

import Grid from '@material-ui/core/Grid';
import Divider from '@material-ui/core/Divider';
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";

import {BountyForm} from "./BountyForm";
import {default as BountyTile} from "./BountyTile";

import {RepositoryStore} from "../stores/RepositoryStore";
import {BountyStore} from "../stores/BountyStore";
import {UIStore} from "../stores/UIStore";

import * as css from './app.scss';

interface Props {
    repoStore?: RepositoryStore;
    bountyStore?: BountyStore;
    uiStore?: UIStore;
}

@inject("uiStore")
@inject("bountyStore")
@inject("repoStore")
@observer
export default class Bounties extends React.Component<Props, {}> {

    componentWillMount() {
        let {repo} = this.props.repoStore;
        this.props.bountyStore.fetchBountiesOfRepo(repo.owner, repo.name);
    }

    componentWillUnmount() {

    }

    toggleNewBountyForm = () => {
        this.props.uiStore.toggleBountyFormOpen();
    }

    render() {
        let {bounties, loading} = this.props.bountyStore;
        let {repo} = this.props.repoStore;
        let {bountyFormOpen} = this.props.uiStore;

        let bountyElements = [];
        bounties.forEach((v, k) => {
            bountyElements.push(<BountyTile key={k} bounty={v} repo={repo}/>);
        });

        return (
            <React.Fragment>
                <Grid container justify="flex-start" spacing={16}>
                    <Grid item xs={12}>
                        <Typography component="h2">
                            Bounties ({bountyElements.length})
                        </Typography>
                        <Divider className={css.dividerSmall}/>
                    </Grid>
                    <Grid item xs={12}>
                        <Typography component="p">
                            Bounties which are linked to issues of this repository.
                        </Typography>
                        <Typography component="p">
                            Balances on pool addresses are updated regularly and might not reflect the correct balance.
                        </Typography>
                    </Grid>
                    <Grid item xs={12} className={css.marginBottom}>
                        <Button className={clsx({[css.marginBottom]: bountyFormOpen})} variant="outlined"
                                color="primary" onClick={this.toggleNewBountyForm}>
                            {bountyFormOpen ? 'Close Form' : 'Add Bounty'}
                        </Button>
                        {bountyFormOpen && <BountyForm/>}
                        {bountyFormOpen && <Divider className={css.dividerSmall}/>}
                    </Grid>
                </Grid>
                <Grid container justify="flex-start" spacing={16}>
                    {bountyElements.length > 0 ?
                        bountyElements
                        :
                        <Grid item>
                            <Typography component="p">
                                No bounties were added yet.
                            </Typography>
                        </Grid>
                    }
                </Grid>
            </React.Fragment>
        );
    }
}