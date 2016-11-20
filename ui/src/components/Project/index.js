import React, { Component } from 'react';
import ProjectInfo from '../ProjectInfo';
import TestCaseUpload from '../TestCaseUpload';
import TestCaseRun from '../TestCaseRun';

import './project.css';

export default class Project extends Component {
  static propTypes = {
    hash: React.PropTypes.string
  };

  // consider caching the selected project in state
  render() {
    let {
      hash
    } = this.props;
    return (
      <div className="project-overview">
        <div className={ `project-info-upload column` }>
          <ProjectInfo hash={ hash } />
          <TestCaseUpload hash={ hash } />
        </div>
        <div className={ `column` }>
          <TestCaseRun hash={ hash } />
        </div>
      </div>
    );
  }
}
