import React from 'react';
import ReactDOM from 'react-dom';
import { Router, Route, hashHistory } from 'react-router'

import App from './components/App';
import Home from './components/Home';

ReactDOM.render((
        <Router history={hashHistory}>
            <Route component={App}>
                <Route path="/" component={Home} />
            </Route>
        </Router>
    ),
    document.getElementById('root')
);
