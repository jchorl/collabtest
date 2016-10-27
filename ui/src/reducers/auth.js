import Immutable from 'immutable';

import { REQUEST_AUTH, RECEIVE_AUTH } from '../actions';

export default function auth(state = Immutable.Map({
    fetched: false,
    authd: false
}), action) {
    switch (action.type) {
        case REQUEST_AUTH:
            return state.set('fetched', false)
        case RECEIVE_AUTH:
            return Immutable.Map({
                fetched: true,
                authd: action.success
            });
        default:
            return state
    }
}
