import { combineReducers } from 'redux';
import { reducer as formReducer } from 'redux-form';
import auth from './auth';
import projects from './projects';
import uploadTestCases from './uploadTestCases';

export default combineReducers({
    auth,
    projects,
    uploadTestCases,
    form: formReducer
})
