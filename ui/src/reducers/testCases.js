import Immutable from 'immutable';

import { REQUEST_TEST_CASES, RECEIVE_TEST_CASES, UPLOAD_TEST_CASE_COMPLETE } from '../actions';

export default function testCases(state = Immutable.Map({
  fetched: false,
  testCases: Immutable.Map()
}), action) {
  switch (action.type) {
    case REQUEST_TEST_CASES:
      return state.setIn([action.hash, 'fetched'], false);
    case RECEIVE_TEST_CASES:
      return state.setIn([action.hash, 'fetched'], true).setIn([action.hash, 'testCases'], Immutable.fromJS(action.testCases));
    case UPLOAD_TEST_CASE_COMPLETE:
      // TODO fix issue with replacing a test case (same content) and having both show up
      return state.setIn([action.hash, 'fetched'], true).setIn([action.hash, 'testCases', Math.random()], Immutable.fromJS(action.testCase));
    default:
      return state
  }
}
