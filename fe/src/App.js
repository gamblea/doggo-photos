import "./App.css";

import { Row, Col, Container, Button } from "react-bootstrap";

import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link,
  useHistory,
} from "react-router-dom";
import { useState, useEffect } from "react";

import "bootstrap/dist/css/bootstrap.min.css";

import Cookies from "js-cookie";

// Internal
import { Login } from "./Login.js";
import { CreateAccount } from "./CreateAccount.js";
import { Dashboard } from "./Dashboard.js";

function App() {
  const [loginKey, setloginKey] = useState("");
  useEffect(() => {
    const key = Cookies.get("doggo-photos-loginKey");
    if (key) {
      setloginKey(key);
    }
  }, []);
  return (
    <Router>
      <Container>
        <Switch>
          <Route exact path="/">
            <StartMenu loginKey={loginKey} setloginKey={setloginKey} />
          </Route>
          <Route exact path="/login">
            <Login setloginKey={setloginKey} />
          </Route>
          <Route exact path="/create">
            <CreateAccount setloginKey={setloginKey} />
          </Route>
          <Route exact path="/dashboard">
            <Dashboard loginKey={loginKey} setloginKey={setloginKey} />
          </Route>
        </Switch>
      </Container>
    </Router>
  );
}

function StartMenu({ loginKey, setloginKey }) {
  const history = useHistory();
  useEffect(() => {
    if (loginKey) {
      history.push("/dashboard");
    } else {
      const cookieToken = Cookies.get("doggo-photos-loginKey");
      if (cookieToken) {
        setloginKey(cookieToken);
        history.push("/dashboard");
      }
    }
  });

  return (
    <Row className="center-block text-center start-menu">
      <Col>
        <h1>Doggo Photos</h1>
        <Link to={`/login`}>
          <Button variant="primary">Login</Button>
        </Link>
        <Link to="/create">
          <Button variant="secondary">Create Account</Button>
        </Link>
      </Col>
    </Row>
  );
}

export default App;
