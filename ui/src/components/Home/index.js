import React, { Component } from 'react';
import { connect } from 'react-redux'
import { mapContains } from 'react-immutable-proptypes';

import './home.css';

class Home extends Component {
    static propTypes = {
        auth: mapContains({
            authd: React.PropTypes.bool.isRequired,
            fetched: React.PropTypes.bool.isRequired
        })
    };

    componentWillReceiveProps(nextProps) {
        if (nextProps.auth.get('authd')) {
            nextProps.history.push('/dashboard');
        }
    }

    render() {
        if (this.props.auth.get('fetched')) {
            return (
                <div className="container-wrapper">
                    <div className="container">
                        <h1 className="welcome">
                            Welcome to CollabTest!
                        </h1>
                        <p>
                            CollabTest is a tool that allows anybody to submit test cases for a project and anybody to run those tests on a provided program. To begin, please
                        </p>
                        <a href="https://github.com/login/oauth/authorize?client_id=47ecbefcf49c1c3ce7d4" className={`github button`}><span className="github-icon"><i className={`fa fa-github`} aria-hidden="true"></i></span>Login with GitHub</a>
                    </div>
                </div>
            );
        }
        return null;
    }
}

export default connect(store => {
    return {
        auth: store.auth
    }
})(Home);
