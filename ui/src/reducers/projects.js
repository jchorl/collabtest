import Immutable from 'immutable';

import { REQUEST_PROJECTS, RECEIVE_PROJECTS, PROJECT_CREATED } from '../actions';

export default function projects(state = Immutable.Map({
    fetched: false,
    fetching: false,
    projects: Immutable.List()
}), action) {
    switch (action.type) {
        case REQUEST_PROJECTS:
            return state.set('fetched', false).set('fetching', true);
        case RECEIVE_PROJECTS:
            return Immutable.Map({
                fetched: true,
                fetching: false,
                projects: Immutable.fromJS(action.projects)
            });
        case PROJECT_CREATED:
            let imm = Immutable.fromJS(action.project);
            return state.update('projects', projects => projects.unshift(imm));
        default:
            return state
    }
}
