import * as React from 'react';
import {inject, observer} from 'mobx-react';
import DevTools from 'mobx-react-devtools';
import {withRouter} from "react-router";
import {Link, Route, Switch} from 'react-router-dom';

import AppBar from '@material-ui/core/AppBar';
import CssBaseline from '@material-ui/core/CssBaseline';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import {withStyles} from '@material-ui/core/styles';

import {UIStore} from "../stores/UIStore";

import * as css from './app.scss';

import Repository from "./Repository";
import Bounty from "./Bounty";
import Repositories from "./Repositories";

declare var __DEVELOPMENT__;

interface Props {
    uiStore?: UIStore;
    classes?: any;
    theme?: any;
}

const drawerWidth = 240;

const styles = theme => ({
    root: {
        display: 'flex',
    },
    appBar: {},
    menuButton: {
        marginRight: 20,
        [theme.breakpoints.up('sm')]: {
            display: 'none',
        },
    },
    toolbar: theme.mixins.toolbar,
    drawerPaper: {
        width: drawerWidth,
    },
});

@withRouter
@inject("uiStore")
@observer
class app extends React.Component<Props, {}> {

    render() {
        const {classes} = this.props;

        return (
            <div className={classes.root}>
                <CssBaseline/>
                <AppBar position="fixed" className={classes.appBar}>
                    <Toolbar>
                        <Link to={"/"} className={css.siteTitle}>
                            <Typography variant="h6" color="inherit" noWrap>
                                IOTA Bounty Platform
                            </Typography>
                        </Link>
                    </Toolbar>
                </AppBar>
                <main className={css.content}>
                    <div className={classes.toolbar}/>
                    <Switch>
                        <Route exact path="/repo/:owner/:name" component={Repository}/>
                        <Route exact path="/bounty/:id" component={Bounty}/>
                        <Route component={Repositories}/>
                    </Switch>
                </main>
                {__DEVELOPMENT__ && <DevTools/>}
            </div>
        );
    }
}

export default withStyles(styles, {withTheme: true})(app);