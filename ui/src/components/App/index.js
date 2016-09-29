import React, { Component } from 'react';
import { connect } from 'react-redux';
import { fetchAuth } from '../../actions';
import Nav from '../Nav';

// global imports for the entire application
import 'font-awesome/css/font-awesome.min.css';

import './app.css';

class App extends Component {
    constructor(props) {
        super(props);
        props.dispatch(fetchAuth());
    }

    render() {
        return (
            <div>
                <Nav />
                <div>
                    {this.props.children}
                </div>
            </div>
        );
    }
}

export default connect()(App);