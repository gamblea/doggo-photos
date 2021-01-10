import { useForm } from "react-hook-form";
import { Row, Col, Button, Form, Alert } from "react-bootstrap";
import axios from "axios";
import { yupResolver } from "@hookform/resolvers/yup";
import * as yup from "yup";
import { useState } from "react";
import { useHistory } from "react-router-dom";
import Cookies from "js-cookie";

function createOnCreateAccount(setloginKey, setCreateAccountError, history) {
  return (data) => {
    axios
      .post(`/api/account/create`, {
        username: data.username,
        password: data.password,
      })
      .then((res) => {
        const loginKey = res.data?.loginKey;
        if (loginKey) {
          Cookies.set("doggo-photos-loginKey", res.data.loginKey, {
            expires: 7,
          });
          setloginKey(loginKey);

          history.push("/dashboard");
        } else {
          const loginError = res.data?.error ? res.data?.error : "Login Error";
          setCreateAccountError(loginError);
        }
      })
      .catch((err) => {
        setCreateAccountError("Could not create account");
      });
  };
}

const schema = yup.object().shape({
  username: yup.string().required("Required"),
  password: yup.string().required("Required"),
  confirmPassword: yup
    .string()
    .oneOf([yup.ref("password"), null], "Passwords do not match!")
    .required("Required"),
});

export function CreateAccount({ setloginKey }) {
  const [createAccountError, setCreateAccountError] = useState("");
  const history = useHistory();
  const { register, handleSubmit, errors } = useForm({
    resolver: yupResolver(schema),
  });
  return (
    <Row className="vertical-center">
      <Col md="3"></Col>
      <Col>
        <Form
          onSubmit={handleSubmit(
            createOnCreateAccount(setloginKey, setCreateAccountError, history)
          )}
        >
          <Form.Group controlId="formBasicEmail">
            <Form.Label>Username</Form.Label>
            <Form.Control
              name="username"
              type="text"
              placeholder="Username"
              ref={register}
            />
          </Form.Group>
          {errors.username && (
            <Alert variant="warning">{errors.username?.message}</Alert>
          )}
          <Form.Group controlId="formBasicPassword">
            <Form.Label>Password</Form.Label>
            <Form.Control
              name="password"
              type="password"
              placeholder="Password"
              ref={register}
            />
          </Form.Group>
          {errors.password && (
            <Alert variant="warning">{errors.password?.message}</Alert>
          )}
          <Form.Group controlId="formBasicPassword">
            <Form.Label>Confirm Password</Form.Label>
            <Form.Control
              name="confirmPassword"
              type="password"
              placeholder="Repeat Password"
              ref={register}
            />
          </Form.Group>
          {errors.confirmPassword && (
            <Alert variant="warning">{errors.confirmPassword?.message}</Alert>
          )}
          <Button variant="primary" type="submit">
            Submit
          </Button>
        </Form>
        {createAccountError ? (
          <div style={{ marginTop: "10px" }}>
            <Alert variant="warning">{createAccountError}</Alert>
          </div>
        ) : null}
      </Col>
      <Col md="3"></Col>
    </Row>
  );
}
