export const REQUEST_AUTH = 'REQUEST_AUTH';
export const RECEIVE_AUTH = 'RECEIVE_AUTH';
export const REQUEST_PROJECTS = 'REQUEST_PROJECTS';
export const RECEIVE_PROJECTS = 'RECEIVE_PROJECTS';
export const PROJECT_CREATED = 'PROJECT_CREATED';

function requestAuth() {
    return { type: REQUEST_AUTH };
}

function receiveAuth(success) {
    return { type: RECEIVE_AUTH, success };
}

export function fetchAuth() {
    return dispatch => {
        dispatch(requestAuth());
        fetch('/api/auth/loggedIn', {
            headers: {
                'Accept': 'application/json'
            },
            credentials: 'include'
        })
        .then(resp => dispatch(receiveAuth(resp.status === 200)));
    }
}

function requestProjects() {
    return { type: REQUEST_PROJECTS };
}

function receiveProjects(projects) {
    return { type: RECEIVE_PROJECTS, projects };
}

export function fetchProjects() {
    return dispatch => {
        dispatch(requestProjects());
        fetch('/api/projects', {
            headers: {
                'Accept': 'application/json'
            },
            credentials: 'include'
        })
            .then(resp => resp.json())
            .then(parsed => dispatch(receiveProjects(parsed)));
    }
}

function projectCreated(project) {
    return { type: PROJECT_CREATED, project };
}

export function createProject(project) {
    return dispatch => {
        fetch("/api/projects", {
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            credentials: 'include',
            method: 'POST',
            body: JSON.stringify(project)
        })
            .then(resp => resp.json())
            .then(parsed => dispatch(projectCreated(parsed)));
    }
}
