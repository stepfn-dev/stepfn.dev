import React, {useState} from "react";
import Button from "react-bootstrap/Button";
import Modal from "react-bootstrap/Modal";
import Alert from "react-bootstrap/Alert";

export function WelcomeButtonAndModal() {
    const hasSeenWelcome = localStorage.getItem("hasSeenWelcome") === "true";
    // const hasSeenWelcome = false;
    const [show, setShow] = useState(!hasSeenWelcome);

    const handleClose = () => {
        localStorage.setItem("hasSeenWelcome", "true");
        setShow(false);
    }
    const handleShow = () => setShow(true);

    return (
        <>
            <a className={"nav-item nav-link topnav-link"} href="#" onClick={handleShow}>Welcome</a>
            <Modal show={show} onHide={handleClose} size={"lg"}>
                <Modal.Header closeButton>
                    <Modal.Title>Welcome to <code>stepfn.dev</code></Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Alert variant={"warning"}>
                        This site is very much a work-in-progress - and I'm seeking
                        help from the open source community! The code is all on GitHub
                        if you want to help out.
                    </Alert>
                    <h3>What</h3>
                    <p>
                        <code>stepfn.dev</code> is intended to be like JSFiddle, JS Bin,
                        et al - but for AWS Step Functions.
                    </p>
                    <h3>How</h3>
                    <p>
                        There are four panels available.
                        </p>
                    <p>
                        Top-left is where you write your state machine definition.
                    </p>
                    <p>
                        Top-right is where you can write Javascript to simulate Lambda
                        functions. <code>Parameter.FunctionName</code> is where you specify
                        the JS function name to execute in a <code>Task</code> state.
                    </p>
                    <p>
                        Bottom-left is where you write input to be passed to an
                        execution of the Step Function.
                    </p>
                    <p>
                        Finally, bottom-right is where output from executions of
                        your Step Function will appear.
                    </p>
                    <p>
                        You can either click the <mark>Execute</mark> button or
                        press <kbd>Cmd</kbd>+<kbd>Enter</kbd> on the keyboard to execute your Step
                        Function.
                    </p>

                    <h3>Why</h3>
                    <p>
                        I built this site for <s>two</s> three reasons:
                    </p>
                    <ol className={"list-group list-group-flush"}>
                        <li className={"list-group-item"}>
                            Iterating on AWS Step Functions felt like it was
                            harder than necessary. You'd need multiple tabs
                            open, create functions, roles, copy and paste ARNs
                            and click a lot of buttons. I thought I could try
                            improve on that experience.
                        </li>
                        <li className={"list-group-item"}>
                            Collaborating on those step functions was tricky,
                            to the point of not bothering when it came to seeking
                            help on Twitter, Stack Overflow, etc. I thought
                            a website like this one might make it easier to share
                            functions for discussion.
                        </li>
                        <li className={"list-group-item"}>
                            I haven't made a frontend website since the days
                            of Microsoft FrontPage and I really needed the
                            practice - and <b>your</b> help to improve it.
                        </li>
                    </ol>
                    <p>
                        This welcome message will only display the first time you visit
                        the site. To see it again, click the <mark>Welcome</mark> button in
                        the nav bar.
                    </p>
                </Modal.Body>
            </Modal>
        </>
    );
}
