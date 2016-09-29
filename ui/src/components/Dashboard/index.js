import React, { Component } from 'react';
import { connect } from 'react-redux'
import { mapContains } from 'react-immutable-proptypes';
import { fetchProjects, createProject } from '../../actions';
import ProjectsBar from '../ProjectsBar';
import CreateProject from '../CreateProject';

import './dashboard.css';

class Dashboard extends Component {
    static propTypes = {
        projects: mapContains({
            fetched: React.PropTypes.bool.isRequired
        }),
        auth: mapContains({
            authd: React.PropTypes.bool.isRequired,
            fetched: React.PropTypes.bool.isRequired
        })
    };

    constructor(props) {
        super(props);
        if (props.auth.get('fetched') && !props.auth.get('authd')) {
            props.history.push('/dashboard');
        } else if (props.auth.get('fetched')) {
            props.dispatch(fetchProjects());
        }

        let selected = '';
        if (props.projects.get('fetched') && props.projects.get('projects').size !== 0) {
            selected = props.projects.get('projects').first().get('hash')
        }

        this.state = {
            selected
        }
    }

    componentWillReceiveProps(nextProps) {
        if (!nextProps.auth.get('authd')) {
            nextProps.history.push('/');
        } else if (!nextProps.projects.get('fetching') && !nextProps.projects.get('fetched')) {
            nextProps.dispatch(fetchProjects());
        } else if (nextProps.projects.get('fetched') && nextProps.projects.get('projects').size !== 0) {
            this.setState({ selected: nextProps.projects.get('projects').first().get('hash') });
        }
    }

    newProject = data => {
        this.props.dispatch(createProject(data));
    }

    onProjectSelect = hash => {
        let f = function() {
            this.setState({ selected: hash });
        }
        return f.bind(this);
    }

    render() {
        const { selected } = this.state;

        return this.props.auth.get('authd') && this.props.projects.get('fetched') ? (
            <div className="dashboard">
                <ProjectsBar selected={ selected } onProjectSelect={ this.onProjectSelect } />
                <div className="dashboard-content-container">
                    <div className="dashboard-content">
                        {
                            selected === '' ? (
                                <CreateProject onSubmit={ this.newProject } />
                            ) : null
                        }
                    </div>
                </div>
            </div>
        ) : null;
    }
}

export default connect(store => {
    return {
        projects: store.projects,
        auth: store.auth,
    }
})(Dashboard);
