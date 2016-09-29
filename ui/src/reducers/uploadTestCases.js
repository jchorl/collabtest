import Immutable from 'immutable';

import { UPLOAD_TEST_CASES, UPLOAD_TEST_CASES_COMPLETE } from '../actions';

export default function auth(state = Immutable.Map({
    success: false
}), action) {
    switch (action.type) {
        case UPLOAD_TEST_CASES:
            return state.set('success', false)
        case UPLOAD_TEST_CASES_COMPLETE:
            return state.set('success', true)
        default:
            return state
    }
}
