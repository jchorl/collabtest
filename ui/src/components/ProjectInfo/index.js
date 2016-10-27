import React, { Component } from 'react';
import { connect } from 'react-redux';
import { list } from 'react-immutable-proptypes';

import './project-info.css';

class ProjectInfo extends Component {
    static propTypes = {
        projects: list,
        hash: React.PropTypes.string.isRequired
    };

    // consider caching the selected project in state
    render() {
        const {
            hash,
            projects
        } = this.props;
        let proj = projects.find(p => p.get('hash') === hash);
        let link = `${location.protocol}//${location.hostname}${location.port ? ':' + location.port : ''}/projects/${proj.get('hash')}`;
        return (
            <div className="project-info">
                <h1>{ proj.get('name') }</h1>
                <div>Link: <a href={link} target="_blank"> {link}</a></div>
                <div>Created: { (new Date(proj.get('createdAt'))).toLocaleString() }</div>
            </div>
        );
    }
}

export default connect(store => {
    return {
        projects: store.projects.get('projects')
    }
})(ProjectInfo);
