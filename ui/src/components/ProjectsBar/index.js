import React, { Component } from 'react';
import { connect } from 'react-redux'
import { list } from 'react-immutable-proptypes';

import './projects-bar.css';

class ProjectsBar extends Component {
    static propTypes = {
        projects: list,
        selected: React.PropTypes.string.isRequired,
        onProjectSelect: React.PropTypes.func
    };

    render() {
        const {
            projects,
            onProjectSelect,
            selected
        } = this.props;

        return (
            <div className="projects-bar">
                <div className={`project` + (this.props.selected === '' ? ` selected` : ``)} onClick={ onProjectSelect('') }>
                    <i className={`fa fa-plus-circle`} aria-hidden="true"></i> New
                </div>
                {
                    projects.map(project => <div key={ project.get('hash') } className={`project` + (selected === project.get('hash') ? ` selected` : ``)} onClick={ onProjectSelect(project.get('hash')) }>{ project.get('name') }</div>)
                }
            </div>
        );
    }
}

export default connect(store => {
    return {
        projects: store.projects.get('projects')
    }
})(ProjectsBar);
