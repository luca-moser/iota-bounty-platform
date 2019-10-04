import * as React from 'react';
import Grid from '@material-ui/core/Grid';
import Card from '@material-ui/core/Card';
import CardActions from '@material-ui/core/CardActions';
import CardContent from '@material-ui/core/CardContent';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';
import LaunchIcon from '@material-ui/icons/Launch';
import {Repository} from "../stores/RepositoryStore";
import {withStyles} from "@material-ui/core";
import {Link} from 'react-router-dom';

const styles = {
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

interface RepoTileProps {
    repo: Repository;
    classes?: any;
}

class repoTile extends React.Component<RepoTileProps, {}> {
    render() {
        const {classes} = this.props;
        let repo = this.props.repo;

        return (
            <Grid item className={css.tile}>
                <Card>
                    <CardContent>
                        <Typography component="h2">
                            {repo.name}
                        </Typography>
                        <Typography className={classes.pos} color="textSecondary">
                            {repo.owner}
                        </Typography>
                        <Typography component="p">
                            {repo.description}
                        </Typography>
                    </CardContent>
                    <CardActions>
                        <Link to={`/repo/${repo.owner}/${repo.name}`}>
                            <Button size="small">
                                <LaunchIcon className={css.marginRight}/>
                                Expand
                            </Button>
                        </Link>
                    </CardActions>
                </Card>
            </Grid>
        );
    }
}

export default withStyles(styles, {withTheme: true})(repoTile);