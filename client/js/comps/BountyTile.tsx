import * as React from 'react';
import {withStyles} from "@material-ui/core";
import {Link} from 'react-router-dom';
import * as dateformat from 'dateformat';

import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';
import LaunchIcon from '@material-ui/icons/Launch';
import Divider from "@material-ui/core/Divider";

import {Bounty, BountyState, mapStateToStr} from "../stores/BountyStore";
import {Repository} from "../stores/RepositoryStore";

import * as css from './app.scss';

const styles = {
    card: {
        minWidth: 275,
    },
    bullet: {
        display: 'inline-block',
        margin: '0 2px',
        transform: 'scale(0.8)',
    },
    title: {
        fontSize: 14,
    },
    pos: {
        marginBottom: 12,
    },
};

interface BountyTileProps {
    bounty: Bounty;
    repo: Repository;
    classes?: any;
}

class bountyTile extends React.Component<BountyTileProps, {}> {
    render() {
        const {classes} = this.props;
        let bounty = this.props.bounty;
        let repo = this.props.repo;

        return (
            <Grid item className={css.tile}>
                <Card className={classes.card}>
                    <CardContent>
                        <Typography component="h2">
                            {bounty.title}
                        </Typography>
                        <Typography className={classes.pos} color="textSecondary">
                            Linked to issue with ID: {bounty.id}
                        </Typography>
                        <Typography component="h3">
                            State
                        </Typography>
                        {
                            bounty.state == BountyState.Transferred ?
                                <div>
                                    <Typography component="p">
                                        {`${mapStateToStr(bounty.state)} `}
                                        <a className={css.underlined} href={`https://thetangle.org/bundle/${bounty.bundle_hash}`} target={'_blank'}>
                                            bundle
                                        </a>
                                        {` to `}
                                        <a className={css.underlined} href={`https://thetangle.org/address/${bounty.receiver_address}`} target={'_blank'}>
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
                        <Typography component="h3">
                            Bounty Address
                        </Typography>
                        <div>
                            <a href={`https://thetangle.org/address/${bounty.pool_address}`} target={'_blank'}>
                                <div className={css.addressBox}>{bounty.pool_address}</div>
                            </a>
                        </div>
                        <Divider className={css.dividerMiddle}/>
                        <Typography component="h3">
                            Balance
                        </Typography>
                        <Typography component="p">
                            {bounty.balance} iotas
                        </Typography>
                        <Divider className={css.dividerMiddle}/>
                        <Typography color="textSecondary">
                            Created on: {dateformat(bounty.created_on, "dd.mm.yyyy HH:MM:ss")}
                            <br/>
                            {
                                bounty.updated_on !== null &&
                                <span>Last updated: {dateformat(bounty.updated_on, "dd.mm.yyyy HH:MM:ss")}</span>
                            }
                        </Typography>
                    </CardContent>
                    <CardActions>
                        <Link to={`/bounty/${bounty.id}`}>
                            <Button size="small" color='primary'>
                                <LaunchIcon className={css.marginRight}/>
                                View
                            </Button>
                        </Link>
                        <a
                            href={`https://github.com/${repo.owner}/${repo.name}/issues/${bounty.issue_number}`}
                            target='_blank'
                        >
                            <Button size="small" color='secondary'>
                                <LaunchIcon className={css.marginRight}/>
                                Issue on GitHub
                            </Button>
                        </a>
                    </CardActions>
                </Card>
            </Grid>
        );
    }
}

export default withStyles(styles, {withTheme: true})(bountyTile);
