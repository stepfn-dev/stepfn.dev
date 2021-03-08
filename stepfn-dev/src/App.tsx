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
import Badge from 'react-bootstrap/Badge';
import {ExclamationCircle, Github} from 'react-bootstrap-icons';
import './App.css';
import {WelcomeButtonAndModal} from "./welcome";

function uuidv4() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

function keyForId(id: string): string|null {
    return localStorage.getItem(`key:${id}`);
}

function App() {
    const splitHash = window.location.hash.split("/sfn/");
    const [id, setId] = useState(splitHash.length == 2 ? splitHash[1] : "");
    // let hashnoop = false;
    // window.addEventListener("hashchange", () => {
    //     if (!hashnoop) {
    //         debugger;
    //         const splitHash = window.location.hash.split("/sfn/");
    //         setId(splitHash.length == 2 ? splitHash[1] : "");
    //     }
    // });

    useEffect(() => {
        // hashnoop = true;
        window.history.replaceState(null, '', `#/sfn/${id}`);
        // hashnoop = false;
    }, [id]);

    const [initialLoad, setInitialLoad] = useState(true);
    const [isLoading, setLoading] = useState(false);
    const [error, setError] = useState(false);

    let v: Values = {
        Definition: "",
        Script: "",
        Input: "",
        Key: uuidv4()
    }
    if (id === "") {
        v = defaultValues();
    }
    const [script, setScript] = useState(v.Script);
    const [definition, setDefinition] = useState(v.Definition);
    const [input, setInput] = useState(v.Input);
    const [output, setOutput] = useState("// Upon execution, the step function's output will appear here");

    useEffect(() => {
        if (initialLoad && id !== "") {
            const get = async() => {
                const resp = await fetch(`https://api.stepfn.dev/sfn?id=${id}`);
                const j = await resp.json();
                setScript(j.Script);
                setDefinition(j.Definition);
                setInput(j.Input);
            }

            get();
            setInitialLoad(false);
        }
    }, [initialLoad]);

    useEffect(() => {
        if (isLoading) {
            const execute = async () => {
                const values: Values = {
                    Script: script,
                    Definition: definition,
                    Input: input,
                    Id: id,
                    Key: keyForId(id) ?? uuidv4()
                };

                const resp = await fetch("https://api.stepfn.dev/execute", {
                    method: "POST",
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(values)
                });

                setLoading(false);

                const j = await resp.json();
                console.log(j);
                try {
                    const executionOutput = JSON.parse(j.output);
                    const t = JSON.stringify(executionOutput.Output, null, 2);
                    setOutput(t);
                    setError(false);

                    const newId = executionOutput.Id;
                    if (newId !== id) {
                        localStorage.setItem(`key:${newId}`, values.Key);
                        setId(newId);
                    }
                } catch {
                    setError(true);
                    setOutput(j.cause);
                }
            }

            execute();
        }
    }, [isLoading, definition, input, script, id]);

    const [showCaveats, setShowCaveats] = useState(false);

    const handleClose = () => setShowCaveats(false);
    const handleShow = () => setShowCaveats(true);

    const defaultOptions = {
        mode: 'javascript',
        theme: 'blackboard',
        lineNumbers: true,
        lineWrapping: true,
        foldGutter: true,
        gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"],
        extraKeys: {
            "Cmd-Enter": () => setLoading(true)
        }
    };
    return (
        <div className="App">
            <Navbar bg={"light"}>
                <Navbar.Brand><code>stepfn.dev</code></Navbar.Brand>
                <Nav className={"mr-auto"}>
                    <Button variant={"outline-primary"}>New Step Function</Button>
                    <Button
                        variant="primary"
                        disabled={isLoading}
                        onClick={!isLoading ? () => setLoading(true) : () => {
                        }}
                    >
                        {isLoading ? 'Executing…' : 'Execute'}
                    </Button>
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
                            <Card.Header>Execution Output {error &&
                            <Badge variant={"danger"}>Error</Badge>}</Card.Header>
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
    Id?: string;
    Key: string;
}

function defaultValues(): Values {
    return {
        Key: uuidv4(),
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
