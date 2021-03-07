import React, {useEffect, useState} from 'react';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/blackboard.css';
import 'codemirror/addon/fold/foldcode'
import 'codemirror/addon/fold/foldgutter.css'
import 'codemirror/addon/fold/foldgutter'
import 'codemirror/addon/fold/brace-fold'
import {Controlled} from 'react-codemirror2';
import 'codemirror/mode/javascript/javascript';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Button from 'react-bootstrap/Button';
import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';
import Card from 'react-bootstrap/Card';
import Modal from 'react-bootstrap/Modal';
import {ExclamationCircle, Github} from 'react-bootstrap-icons';
import './App.css';
import {WelcomeButtonAndModal} from "./welcome";

function StartExecutionButton({execute}: { execute(): Promise<any> }) {
    const [isLoading, setLoading] = useState(false);

    useEffect(() => {
        if (isLoading) {
            execute().finally(() => {
                setLoading(false);
            });
        }
    }, [isLoading]);

    const handleClick = () => setLoading(true);

    return (
        <Button
            variant="primary"
            disabled={isLoading}
            onClick={!isLoading ? handleClick : () => {
            }}
        >
            {isLoading ? 'Executing…' : 'Execute'}
        </Button>
    );
}

function App() {
    const v = getValues();

    const [script, setScript] = useState(v.Script);
    const [definition, setDefinition] = useState(v.Definition);
    const [input, setInput] = useState(v.Input);
    const [output, setOutput] = useState("// Upon execution, the step function's output will appear here");

    useEffect(() => {
        const values: Values = {
            Script: script,
            Definition: definition,
            Input: input,
        };

        localStorage.setItem("stepfn-dev-values", JSON.stringify(values));
    }, [script, definition, input]);

    const execute = async () => {
        const values: Values = {
            Script: script,
            Definition: definition,
            Input: input,
        };

        const resp = await fetch("https://api.stepfn.dev/sfn", {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(values)
        });

        const j = await resp.json();
        console.log(j);
        try {
            const t = JSON.stringify(JSON.parse(j.output), null, 2);
            setOutput(t);
        } catch {
            setOutput(j.cause);
        }
    }

    const [showCaveats, setShowCaveats] = useState(false);

    const handleClose = () => setShowCaveats(false);
    const handleShow = () => setShowCaveats(true);

    const defaultOptions = {
        mode: 'javascript',
        theme: 'blackboard',
        lineNumbers: true,
        foldGutter: true,
        gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"]
    };
    return (
        <div className="App">
            <Navbar bg={"light"}>
                <Navbar.Brand><code>stepfn.dev</code></Navbar.Brand>
                <Nav className={"mr-auto"}>
                    <Button variant={"outline-primary"}>New Step Function</Button>
                    <StartExecutionButton execute={execute}/>
                    <Button variant={"outline-info"}>Share…</Button>
                </Nav>
                <Nav>
                    <WelcomeButtonAndModal/>
                    <Button variant={"outline-danger"} onClick={handleShow}>Known Issues <ExclamationCircle/></Button>
                    <Button href={"https://github.com/stepfn-dev/stepfn.dev"} variant={"outline-info"}>GitHub <Github/></Button>
                </Nav>
            </Navbar>
            <Modal show={showCaveats} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>Known Issues <ExclamationCircle/></Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <ul className={"list-group list-group-flush"}>
                        <li className={"list-group-item"}>
                            Only <code>arn:aws:states:::lambda:invoke</code> resources are supported,
                            i.e. <code>"Resource": "&lt;lambda function arn&gt;"</code> is not supported.
                        </li>
                        <li className={"list-group-item"}>
                            Error handling is mostly non-existent right now.
                        </li>
                        <li className={"list-group-item"}>
                            I need a lot of help with the frontend, in case you hadn't noticed.
                        </li>
                    </ul>
                </Modal.Body>
            </Modal>
            <Container fluid>
                <Row className={"mt-3"}>
                    <Col>
                        <Card>
                            <Card.Header>Definition</Card.Header>
                            <Controlled
                                className={"editor"}
                                value={definition}
                                onBeforeChange={(editor, data, value) => {
                                    setDefinition(value);
                                }}
                                options={defaultOptions}
                                onChange={(editor, data, value) => {
                                }}
                            />
                        </Card>
                    </Col>
                    <Col>
                        <Card>
                            <Card.Header>Script</Card.Header>
                            <Controlled
                                className={"editor"}
                                value={script}
                                options={defaultOptions}
                                onBeforeChange={(editor, data, value) => {
                                    setScript(value);
                                }}
                                onChange={(editor, data, value) => {
                                }}
                            />
                        </Card>
                    </Col>
                </Row>
                <Row className={"mt-3 bottom"}>
                    <Col>
                        <Card>
                            <Card.Header>Execution Input</Card.Header>
                            <Controlled
                                className={"editor"}
                                value={input}
                                options={defaultOptions}
                                onBeforeChange={(editor, data, value) => {
                                    setInput(value);
                                }}
                                onChange={(editor, data, value) => {
                                }}
                            />
                        </Card>
                    </Col>
                    <Col>
                        <Card>
                            <Card.Header>Execution Output</Card.Header>
                            <Controlled
                                className={"editor"}
                                value={output}
                                options={defaultOptions}
                                onBeforeChange={(editor, data, value) => {
                                    setOutput(value);
                                }}
                                onChange={(editor, data, value) => {
                                }}
                            />
                        </Card>
                    </Col>
                </Row>
            </Container>
        </div>
    );
}

interface Values {
    Script: string;
    Definition: string;
    Input: string;
}

function getValues(): Values {
    const v = localStorage.getItem("stepfn-dev-values");
    if (v != null) {
        return JSON.parse(v);
    } else {
        const values = defaultValues();
        localStorage.setItem("stepfn-dev-values", JSON.stringify(values));
        return values;
    }
}

function defaultValues(): Values {
    return {
        Input: `{"a": 55, "b": 66}`,
        Script: `
// This function is referenced in the definition on the left using "FunctionName": "sum"
const sum = input => input.First + input.Second;

// Likewise, this function is referenced in the definition on the left using "FunctionName": "unix"
function unix(input) {
    return Date.now();
}        
        `,
        Definition: `
{
  "StartAt": "First Unix date",
  "States": {
    "First Unix date": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "unix"
      },
      "ResultSelector": {
        "First.$": "$.Payload"
      },
      "Next": "Second Unix date"
    },
    "Second Unix date": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "unix"
      },
      "ResultPath": "$.Second",
      "Next": "Refactor"
    },
    "Refactor": {
      "Type": "Pass",
      "Parameters": {
        "First.$": "$.First",
        "Second.$": "$.Second.Payload"
      },
      "Next": "Sum them"
    },
    "Sum them": {
      "Type": "Task",
      "Resource": "arn:aws:states:::lambda:invoke",
      "Parameters": {
        "Payload.$": "$",
        "FunctionName": "sum"
      },
      "ResultSelector": {
        "Sum.$": "$.Payload"
      },
      "ResultPath": "$.Sum",
      "End": true
    }
  }
}        
        `
    }
}

export default App;
