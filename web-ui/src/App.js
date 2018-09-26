import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import {Badge, Col, Row, Collapsible, CollapsibleItem, ProgressBar, Button} from 'react-materialize';

class Step extends Component {
  constructor(props) {
    super(props);
    this.handleSelect = this.handleSelect.bind(this);
  }

  handleSelect(key) {
    const { onSelect } = this.props;

    if (onSelect) { onSelect(key); }

    // if (this.state.activeKey === key) { key = null; }

    if (this.props.accordion) {
      this.setState({ activeKey: key });
    }
  }

  itemHeader(step) {
    return <span>{step.title}<Badge>{step.ok ? "OK" : (step.error ? "Not installed" : "Wrong version") }</Badge></span>;
  }

  render() {
    return <CollapsibleItem className="step" header={ this.itemHeader(this.props.step) } onSelect={this.handleSelect}>
      <p>{this.props.step.description}</p>
      <h3>How to test it</h3>
      <p><tt>{this.props.step.test}</tt></p>
      <h4>What you should get</h4>
      <p><tt>{this.props.step.expected}</tt></p>
      <div>
        <h4>What you actually get</h4>
        <p><tt>{this.props.step.actual ? this.props.step.actual : <i>Nothing</i>}</tt></p>
      </div>
      { this.props.step.error ? 
        <div>
          <h4>This might be why</h4>
          <p><tt>{this.props.step.error}</tt></p>
        </div>
      : "" }
      <h3>How to install it</h3>
      <p><tt>{this.props.step.install}</tt></p>
    </CollapsibleItem>
  }
}

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      doc: {
        title: "",
        steps: [],
        total_steps: 0,
        completed_steps: 0
      },
      loading: false
    };
    this.handleReload = this.handleReload.bind(this);
    this.fetchStepData = this.fetchStepData.bind(this);
  }

  fetchStepData() {
    this.setState({loading:true});
    fetch("http://localhost:5555/steps")
    .then(r => r.json())
    .then(json => {
      console.log(json);
      this.setState({doc: json, loading:false});
    })
    .catch(e => {
      console.log(e);
    });
  }

  handleReload() {
    this.setState({loading:true});
    fetch("http://localhost:5555/reload", {
      method: "POST"
    })
    .then(r => {
      this.fetchStepData();
    })
    .catch(e => {
      console.log(e);
    });
  }
  
  componentDidMount() {
    this.fetchStepData();
  }

  render() {
    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <h1 className="App-title">{ this.state.doc.title }</h1>
          <div>
            <Row>
              <Col s={3} />
              <Col s={6}>
                <ProgressBar progress={ this.state.doc.completed_steps / this.state.doc.total_steps * 100 } />
              </Col>
              <Col s={3} />  
            </Row>
            <p>
              { this.state.doc.completed_steps } of { this.state.doc.total_steps } steps completed
            </p>
            <div>
              <Button disabled={this.state.loading} onClick={this.handleReload}>Reload</Button>
            </div>
          </div>
        </header>
        <Row>
          <Col s={2} />
          <Col s={8}>
            <Collapsible accordion={true}>
            {
              this.state.doc.steps.map(k => {
                return <Step step={k} />
              })
            }
            </Collapsible>
          </Col>
          <Col s={2} />
        </Row>
      </div>
    );
  }
}

export default App;
