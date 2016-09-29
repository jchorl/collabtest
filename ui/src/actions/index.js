export const REQUEST_AUTH = 'REQUEST_AUTH';
export const RECEIVE_AUTH = 'RECEIVE_AUTH';
export const REQUEST_PROJECTS = 'REQUEST_PROJECTS';
export const RECEIVE_PROJECTS = 'RECEIVE_PROJECTS';
export const PROJECT_CREATED = 'PROJECT_CREATED';
export const UPLOAD_TEST_CASES = 'UPLOAD_TEST_CASES';
export const UPLOAD_TEST_CASES_COMPLETE = 'UPLOAD_TEST_CASES_COMPLETE';
export const RUN_TEST_CASES = 'RUN_TEST_CASES';
export const RUN_TEST_CASES_COMPLETE = 'RUN_TEST_CASES_COMPLETE';

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

function uploadTestCasesCompleted() {
    return {
        type: UPLOAD_TEST_CASES_COMPLETE
    }
}

function beginUploadTestCases() {
    return {
        type: UPLOAD_TEST_CASES
    }
}

export function uploadTestCases(hash, files) {
    return dispatch => {
        dispatch(beginUploadTestCases());
        let data = new FormData();
        data.append('inFile', files[0]);
        data.append('outFile', files[1]);

        fetch(`/api/projects/${hash}/add`, {
            credentials: 'include',
            method: 'POST',
            body: data
        })
        .then(resp => {
            if (resp.status === 202) {
                dispatch(uploadTestCasesCompleted());
            }
        });
    }
}

function runTestCasesCompleted(hash, results) {
    return {
        type: RUN_TEST_CASES_COMPLETE,
        hash,
        results
    }
}

function beginRunTestCases() {
    return {
        type: RUN_TEST_CASES
    }
}

export function runTestCases(hash, file) {
    return dispatch => {
        dispatch(beginRunTestCases());
        let data = new FormData();
        data.append('file', file);

        fetch(`/api/projects/${hash}/run`, {
            credentials: 'include',
            method: 'POST',
            body: data
        })
        .then(resp => resp.json())
        .then(parsed => dispatch(runTestCasesCompleted(hash, parsed)));
    }
}
