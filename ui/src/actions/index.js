export const REQUEST_AUTH = 'REQUEST_AUTH';
export const RECEIVE_AUTH = 'RECEIVE_AUTH';
export const REQUEST_PROJECTS = 'REQUEST_PROJECTS';
export const RECEIVE_PROJECTS = 'RECEIVE_PROJECTS';
export const REQUEST_PROJECT = 'REQUEST_PROJECT';
export const REQUEST_TEST_CASES = 'REQUEST_TEST_CASES';
export const RECEIVE_TEST_CASES = 'RECEIVE_TEST_CASES';
export const PROJECT_CREATED = 'PROJECT_CREATED';
export const UPLOAD_TEST_CASE = 'UPLOAD_TEST_CASE';
export const UPLOAD_TEST_CASE_COMPLETE = 'UPLOAD_TEST_CASE_COMPLETE';
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

function requestProject() {
  return { type: REQUEST_PROJECT };
}

function receiveProject(project) {
  return { type: RECEIVE_PROJECTS, projects: [project] };
}

export function fetchProject(hash) {
  return dispatch => {
    dispatch(requestProject());
    fetch(`/api/projects/${hash}`, {
      headers: {
        'Accept': 'application/json'
      }
    })
      .then(resp => resp.json())
      .then(parsed => dispatch(receiveProject(parsed)));
  }
}

function requestTestCases(hash) {
  return { type: REQUEST_TEST_CASES, hash };
}

function receiveTestCases(hash, testCases) {
  return { type: RECEIVE_TEST_CASES, hash, testCases };
}

export function fetchTestCases(hash) {
  return dispatch => {
    dispatch(requestTestCases(hash));
    fetch(`/api/projects/${hash}/testcases`, {
      headers: {
        'Accept': 'application/json'
      }
    })
      .then(resp => resp.json())
      .then(parsed => dispatch(receiveTestCases(hash, parsed)));
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

function uploadTestCaseCompleted(hash, testCase) {
  return {
    type: UPLOAD_TEST_CASE_COMPLETE,
    hash,
    testCase
  }
}

function beginUploadTestCase() {
  return {
    type: UPLOAD_TEST_CASE
  }
}

export function uploadTestCase(hash, input, output) {
  return dispatch => {
    dispatch(beginUploadTestCase());
    let data = new FormData();
    data.append('inFile', input);
    data.append('outFile', output);

    fetch(`/api/projects/${hash}/add`, {
      credentials: 'include',
      method: 'POST',
      body: data
    })
      .then(resp => resp.json())
      .then(json => {
        dispatch(uploadTestCaseCompleted(hash, json));
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

function beginRunTestCases(hash) {
  return {
    type: RUN_TEST_CASES,
    hash
  }
}

export function runTestCases(hash, file) {
  return dispatch => {
    dispatch(beginRunTestCases(hash));
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
