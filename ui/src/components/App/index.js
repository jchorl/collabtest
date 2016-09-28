import React, { Component } from 'react';
import Nav from '../Nav';

import './app.css';

class App extends Component {
  render() {
    return (
      <div>
            <Nav></Nav>
            <div>
                {this.props.children}
            </div>
      </div>
    );
  }
}

export default App;