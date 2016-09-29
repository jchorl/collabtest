import React, { Component } from 'react';
import Immutable from 'immutable';
import { map } from 'react-immutable-proptypes';
import { connect } from 'react-redux';
import Dropzone from 'react-dropzone';
import { runTestCases } from '../../actions';

import './test-case-run.css';

class TestCaseRun extends Component {
    static propTypes = {
        hash: React.PropTypes.string.isRequired,
        runTestCases: map
    };

    componentWillReceiveProps(nextProps) {
        if (!this.props.runTestCases.get('success') && nextProps.runTestCases.get('success')) {
            alert('Success!');
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
            runTestCases
        } = this.props;

        let results = Immutable.List();
        if (runTestCases.get('lastRun').has(hash)) {
            results = runTestCases.get('lastRun').get(hash).map((diffHtml, idx) => {
                let html = {__html: diffHtml }
                return (
                    <div key={ idx }>
                        <h4>Test {idx}</h4>
                        <div dangerouslySetInnerHTML={ html } />
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
                    runTestCases.get('lastRun').has(hash) ? (
                        <div>
                            <h3>Last run</h3>
                            { results }
                        </div>
                    ) : null
                }
            </div>
        );
    }
}

export default connect(store => {
    return {
        runTestCases: store.runTestCases
    }
})(TestCaseRun);
