import Immutable from 'immutable';

import { UPLOAD_TEST_CASE, UPLOAD_TEST_CASE_COMPLETE } from '../actions';

export default function uploadTestCase(state = Immutable.Map({
    success: false
}), action) {
    switch (action.type) {
        case UPLOAD_TEST_CASE:
            return state.set('success', false)
        case UPLOAD_TEST_CASE_COMPLETE:
            return state.set('success', true)
        default:
            return state
    }
}
