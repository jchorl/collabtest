import React, { Component } from 'react';
import { connect } from 'react-redux';
import { list } from 'react-immutable-proptypes';
import { fetchProject } from '../../actions';
import Project from '../Project';

class ProjectPage extends Component {
  static propTypes = {
    params: React.PropTypes.object.isRequired,
    projects: list.isRequired
  };

  componentDidMount() {
    this.props.dispatch(fetchProject(this.props.params.hash));
  }

  // consider caching the selected project in state
  render() {
    const {
      params: { hash },
      projects
    } = this.props;

    let proj = projects.find(p => p.get('hash') === hash);
    return proj ? (
      <Project hash={ hash } />
    ) : null;
  }
}

export default connect(store => {
  return {
    projects: store.projects.get('projects')
  }
})(ProjectPage);
