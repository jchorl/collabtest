import Immutable from 'immutable';

import { RUN_TEST_CASES, RUN_TEST_CASES_COMPLETE } from '../actions';

export default function runTestCases(state = Immutable.Map(), action) {
  switch (action.type) {
    case RUN_TEST_CASES:
      return state.setIn([action.hash, 'complete'], false);
    case RUN_TEST_CASES_COMPLETE:
      return state.setIn([action.hash, 'complete'], true).setIn([action.hash, 'results'], Immutable.fromJS(action.results));
    default:
      return state
  }
}
