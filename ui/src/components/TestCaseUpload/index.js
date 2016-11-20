import React, { Component } from 'react';
import { map } from 'react-immutable-proptypes';
import { connect } from 'react-redux';
import Dropzone from 'react-dropzone';
import { uploadTestCase } from '../../actions';

import './test-case-upload.css';

class TestCaseUpload extends Component {
  static propTypes = {
    hash: React.PropTypes.string.isRequired,
    uploadTestCase: map
  };

  constructor(props) {
    super(props);
    this.state = { input: null, output: null };
  }

  componentWillReceiveProps(nextProps) {
    if (!this.props.uploadTestCase.get('success') && nextProps.uploadTestCase.get('success')) {
      alert('Success!');
      this.setState({ input: null, output: null });
    }
  }

  onDropInput = files => {
    const {
      hash,
      dispatch
    } = this.props;

    if (this.state.output) {
      dispatch(uploadTestCase(hash, files[0], this.state.output));
    } else {
      this.setState({ input: files[0] });
    }
  }

  onDropOutput = files => {
    const {
      hash,
      dispatch
    } = this.props;

    if (this.state.input) {
      dispatch(uploadTestCase(hash, this.state.input, files[0]));
    } else {
      this.setState({ output: files[0] });
    }
  }

  reset = field => {
    let that = this;

    return () => {
      let newState = {};
      newState[field] = null;
      that.setState(newState);
    }
  }

  render() {
    return (
      <div className="test-case-upload">
        <h2>Submit new test case</h2>
        { this.state.input ? (
          <div>
            <div>Input: { this.state.input.name } <i className={`fa fa-check`} aria-hidden="true"></i></div>
            <button onClick={ this.reset('input') }>Reset</button>
          </div>
        ) : (
          <Dropzone className="dropzone" onDrop={ this.onDropInput } multiple={ false }>
            <div>Please drop input file here</div>
          </Dropzone>
        ) }
        { this.state.output ? (
          <div>
            <div>output: { this.state.output.name } <i className={`fa fa-check`} aria-hidden="true"></i></div>
            <button onClick={ this.reset('output') }>Reset</button>
          </div>
        ) : (
          <Dropzone className="dropzone" onDrop={ this.onDropOutput } multiple={ false }>
            <div>Please drop output file here</div>
          </Dropzone>
        ) }
      </div>
    );
  }
}

export default connect(store => {
  return {
    uploadTestCase: store.uploadTestCase
  }
})(TestCaseUpload);
