import React, { Component } from 'react';
import Immutable from 'immutable';
import { map } from 'react-immutable-proptypes';
import { connect } from 'react-redux';
import Dropzone from 'react-dropzone';
import { runTestCases, fetchTestCases } from '../../actions';

import './test-case-run.css';

class TestCaseRun extends Component {
  static propTypes = {
    hash: React.PropTypes.string.isRequired,
    runTestCases: map,
    testCases: map,
    dispatch: React.PropTypes.func.isRequired
  };

  componentDidMount() {
    this.props.dispatch(fetchTestCases(this.props.hash));
  }

  componentWillReceiveProps(nextProps) {
    if (!this.props.runTestCases.get('success') && nextProps.runTestCases.get('success')) {
      alert('Success!');
    }

    if (nextProps.hash !== this.props.hash) {
      this.props.dispatch(fetchTestCases(nextProps.hash));
    }
  }

  onDrop = files => {
    const {
      hash,
      dispatch
    } = this.props;

    dispatch(runTestCases(hash, files[0]));
  }

  // TODO loop through last runs
  render() {
    const {
      hash,
      runTestCases,
      testCases
    } = this.props;

    let results = Immutable.List();
    let testcases = Immutable.List();
    if (runTestCases.has(hash) && runTestCases.getIn([hash, 'complete'])) {
      results = runTestCases.get(hash).get('results').map((diffHtml, idx) => {
        let html = {__html: diffHtml }
        return (
          <div key={ idx }>
            <h4>Test {idx}</h4>
            <div dangerouslySetInnerHTML={ html } />
          </div>
        )
      });
    } else if (testCases.has(hash) && testCases.getIn([hash, 'fetched'])) {
      testcases = testCases.getIn([hash, 'testCases']).valueSeq().map((testcase, idx) => {
        return (
          <div key={ idx }>
            <h4>Test { idx }:</h4>
            <div><a href={ testcase.get('inputLink') } target="_blank">Input</a> <a href={ testcase.get('outputLink') } target="_blank">Output</a></div>
          </div>
        )
      });
    }

    return (
      <div className="test-case-run">
        <h2>Run your program against test cases</h2>
        <Dropzone className="dropzone" onDrop={ this.onDrop }>
          <div>Please drop application file here</div>
        </Dropzone>
        {
          runTestCases.has(hash) ? (
            <div>
              <h3>Last Run</h3>
              { results }
            </div>
          ) : (
            <div>
              <h3>Test Cases</h3>
              { testCases.getIn([hash, 'fetched']) && testcases.isEmpty() ? <span>No test cases yet</span> : testcases }
            </div>
          )
        }
      </div>
    );
  }
}

export default connect(store => {
  return {
    runTestCases: store.runTestCases,
    testCases: store.testCases
  }
})(TestCaseRun);
