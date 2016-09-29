import React, { Component } from 'react';
import { map } from 'react-immutable-proptypes';
import { connect } from 'react-redux';
import Dropzone from 'react-dropzone';
import { uploadTestCases } from '../../actions';

import './test-case-upload.css';

class TestCaseUpload extends Component {
    static propTypes = {
        hash: React.PropTypes.string.isRequired,
        testCaseUpload: map
    };

    constructor(props) {
        super(props);
        this.state = { files: [] };
    }

    componentWillReceiveProps(nextProps) {
        // TODO should keep track of what got uploaded, multiple uploads, etc
        if (!this.props.uploadTestCases.get('success') && nextProps.uploadTestCases.get('success')) {
            alert('Success!');
            this.setState({ files: [] });
        }
    }

    onDrop = files => {
        const {
            hash,
            dispatch
        } = this.props;

        // TODO no
        if (this.state.files.length === 1) {
            dispatch(uploadTestCases(hash, this.state.files.concat(files)));
        } else {
            this.setState({ files: this.state.files.concat(files) });
        }
    }

    // TODO consider caching the selected project in state
    // TODO make this not terrible
    render() {
        return (
            <div className="test-case-upload">
                <h2>Submit new test case</h2>
                <Dropzone className="dropzone" onDrop={ this.onDrop }>
                    <div>Please drop input file then output file here</div>
                </Dropzone>
            </div>
        );
    }
}

export default connect(store => {
    return {
        uploadTestCases: store.uploadTestCases
    }
})(TestCaseUpload);
