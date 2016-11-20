import { combineReducers } from 'redux';
import { reducer as formReducer } from 'redux-form';
import auth from './auth';
import projects from './projects';
import uploadTestCase from './uploadTestCase';
import runTestCases from './runTestCases';
import testCases from './testCases';

export default combineReducers({
  auth,
  projects,
  testCases,
  uploadTestCase,
  runTestCases,
  form: formReducer
})
