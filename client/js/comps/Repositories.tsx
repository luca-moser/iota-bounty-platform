import * as React from 'react';
import {inject, observer} from 'mobx-react';
import Grid from '@material-ui/core/Grid';

import {UserStore} from "../stores/UserStore";


import * as css from './app.scss';
import {RepositoryStore} from "../stores/RepositoryStore";
import {default as RepoTile} from "./RepoTile";
import Divider from '@material-ui/core/Divider';
import Button from "@material-ui/core/Button";
import Typography from "@material-ui/core/Typography";
import {RepositoryForm} from "./RepositoryForm";
import {UIStore} from "../stores/UIStore";
import clsx from "clsx";

interface Props {
    userStore?: UserStore;
    repoStore?: RepositoryStore;
    uiStore?: UIStore;
}

@inject("uiStore")
@inject("userStore")
@inject("repoStore")
@observer
export default class Repositories extends React.Component<Props, {}> {

    componentWillMount() {
        this.props.repoStore.fetchRepos();
    }

    componentWillUnmount() {

    }

    toggleNewRepoForm = () => {
        this.props.uiStore.toggleRepoFormOpen();
    }

    render() {
        let {repositories} = this.props.repoStore;
        let {repoFormOpen} = this.props.uiStore;

        let repoElements = [];
        repositories.forEach((v, k) => {
            repoElements.push(<RepoTile key={k} repo={v}/>);
        });

        return (
            <React.Fragment>
                <Grid container justify="flex-start" spacing={16}>
                    <Grid item xs={12}>
                        <Typography component="h2">
                            Repositories ({repoElements.length})
                        </Typography>
                        <Divider className={css.dividerSmall}/>
                    </Grid>
                    <Grid item xs={12}>
                        <Typography component="p">
                            Active repositories known to the bounty platform.
                        </Typography>
                    </Grid>
                    <Grid item xs={12} className={css.marginBottom}>
                        <Button className={clsx({[css.marginBottom]: repoFormOpen})} variant="outlined"
                                color="primary" onClick={this.toggleNewRepoForm}>
                            {repoFormOpen ? 'Close Form' : 'Add Repository'}
                        </Button>
                        {repoFormOpen && <RepositoryForm/>}
                        {repoFormOpen && <Divider className={css.dividerSmall}/>}
                    </Grid>
                </Grid>
                <Grid container justify="flex-start" spacing={16}>
                    {repoElements.length > 0 ?
                        repoElements
                        :
                        <Grid item>
                            <Typography component="p">
                                No repositories were added to the bounty platform yet :(
                            </Typography>
                        </Grid>
                    }
                </Grid>
            </React.Fragment>
        );
    }
}