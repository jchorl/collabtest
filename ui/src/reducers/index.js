import { combineReducers } from 'redux';
import { reducer as formReducer } from 'redux-form';
import auth from './auth';
import projects from './projects';
import uploadTestCase from './uploadTestCase';
import runTestCases from './runTestCases';

export default combineReducers({
    auth,
    projects,
    uploadTestCase,
    runTestCases,
    form: formReducer
})
