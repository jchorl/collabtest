import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import { createStore, applyMiddleware } from 'redux';
import thunk from 'redux-thunk';
import { Router, Route, hashHistory } from 'react-router';

import App from './components/App';
import Home from './components/Home';
import Dashboard from './components/Dashboard';
import reducers from './reducers';

let store = createStore(
    reducers,
    applyMiddleware(thunk)
);

ReactDOM.render((
        <Provider store={store}>
            <Router history={hashHistory}>
                <Route component={App}>
                    <Route path="/" component={Home} />
                    <Route path="/dashboard" component={Dashboard} />
                </Route>
            </Router>
        </Provider>
    ),
    document.getElementById('root')
);
