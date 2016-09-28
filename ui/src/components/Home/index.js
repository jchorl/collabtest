import React, { Component } from 'react';
import 'font-awesome/css/font-awesome.min.css';

import './home.css';

class Home extends Component {
  render() {
    return (
        <div className="container-wrapper">
            <div className="container">
                <div className="welcome">
                    Welcome to CollabTest!
                </div>
                <p>
                    CollabTest is a tool that allows anybody to submit test cases for a project and anybody to run those tests on a provided program. To begin, please
                </p>
                <a href="https://github.com/login/oauth/authorize?client_id=47ecbefcf49c1c3ce7d4" className={`github button`}><span className="github-icon"><i className={`fa fa-github`} aria-hidden="true"></i></span>Login with GitHub</a>
            </div>
        </div>
    );
  }
}

export default Home;
