import React, { Component } from 'react';
import { Field, reduxForm } from 'redux-form';

import './create-project.css';

class CreateProject extends Component {
    static propTypes = {
        handleSubmit: React.PropTypes.func.isRequired
    };

    render() {
        const { handleSubmit } = this.props;
        return (
            <div className="create-project">
                <div className="title-wrapper">
                    <h1>Create New Project</h1>
                </div>
                <form onSubmit={ handleSubmit }>
                    <div>
                        <label htmlFor="name">Project Name: </label>
                        <Field name="name" component="input" type="text" />
                    </div>
                    <div>
                        <button type="submit">Submit</button>
                    </div>
                </form>
            </div>
        );
    }
}

export default reduxForm({
    form: 'createProject'
})(CreateProject);
