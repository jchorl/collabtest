import Immutable from 'immutable';

import { RUN_TEST_CASES, RUN_TEST_CASES_COMPLETE } from '../actions';

export default function runTestCases(state = Immutable.Map({
    success: false,
    lastRun: Immutable.Map()
}), action) {
    switch (action.type) {
        case RUN_TEST_CASES:
            return state.set('success', false);
        case RUN_TEST_CASES_COMPLETE:
            return state.set('success', true).setIn(['lastRun', action.hash], Immutable.fromJS(action.results));
        default:
            return state
    }
}
