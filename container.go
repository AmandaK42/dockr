package dockr

import (
  "io"
  "fmt"
  "net/url"
  "encoding/json"
)

/*
The following structs are @authored by docker.
*/

type CreateContainerRequest struct {
  Hostname        string
  User            string
  Memory          int64 // Memory limit (in bytes)
  MemorySwap      int64 // Total memory usage (memory + swap); set `-1' to disable swap
  CpuShares       int64 // CPU shares (relative weight vs. other containers)
  AttachStdin     bool
  AttachStdout    bool
  AttachStderr    bool
  PortSpecs       []string
  Tty             bool // Attach standard streams to a tty, including stdin if it is not closed.
  OpenStdin       bool // Open stdin
  StdinOnce       bool // If true, close stdin after the 1 attached client disconnects.
  Env             []string
  Cmd             []string
  Dns             []string
  Image           string // Name of the image as it was passed by the operator (eg. could be symbolic)
  Volumes         map[string]struct{}
  VolumesFrom     string
  WorkingDir      string
  Entrypoint      []string
  NetworkDisabled bool
  Privileged      bool
}

type CreateContainerResponse struct {
  Id              string
  Warnings        []string
}

type KeyValuePair struct {
  Key string
  Value string
}

type StartContainerRequest struct {
  Binds           []string
  ContainerIDFile string
  LxcConf         []KeyValuePair
}

type StopContainerRequest struct {
  Timeout         int
}

func (q *StopContainerRequest) Values() url.Values {
  query := url.Values{}
  query.Set("t", fmt.Sprintf("%d",q.Timeout))
  return query
}

type AttachContainerRequest struct {
  // Which stream to attach:
  Stdin           bool
  Stdout          bool
  Stderr          bool
  // What to post on this streams:
  Logs            bool // get archived stuff?
  Stream          bool // stream new stuff?
}

func boolString(b bool) string {
  if b {
    return "true"
  }else{
    return "false"
  }
}

func (q *AttachContainerRequest) Values() url.Values {
  query := url.Values{}
  query.Set("stdin" , boolString(q.Stdin) )
  query.Set("stdout", boolString(q.Stdout) )
  query.Set("stderr", boolString(q.Stderr))
  query.Set("logs"  , boolString(q.Logs))
  query.Set("stream", boolString(q.Stream))
  return query
}

func (c *Client) CreateContainer(q *CreateContainerRequest) (*CreateContainerResponse, error){
  res, err := c.callfjson("POST","/v1.4/containers/create",q)
  if err != nil {
    return nil, err
  }
  err = expectHTTPStatus( res.StatusCode, 201 )
  if err != nil {
    return nil, err
  }
  var a CreateContainerResponse
  err = json.NewDecoder(res.Body).Decode(&a)
  if err != nil {
    return nil, err
  }
  return &a, nil
}

func (c *Client) DeleteContainer(id string) error {
  err := validateId(id)
  if err != nil {
    return err
  }
  res, err := c.callf("DELETE","/v1.4/containers/%s",id)
  if err != nil {
    return err
  }
  // 406 = you have to stop before delete
  return expectHTTPStatus(res.StatusCode, 204)
}

func (c *Client) StartContainer(id string, q *StartContainerRequest) error {
  err := validateId(id)
  if err != nil {
    return err
  }
  res, err := c.callfjson("POST","/v1.4/containers/%s/start",q, id)
  if err != nil {
    return err
  }
  return expectHTTPStatus(res.StatusCode, 204)
}

func (c *Client) StopContainer(id string, q *StopContainerRequest) error {
  err := validateId(id)
  if err != nil {
    return err
  }
  res, err := c.callfquery("POST","/v1.4/containers/%s/stop",q.Values(), id)
  if err != nil {
    return err
  }
  return expectHTTPStatus(res.StatusCode, 204)
}
func (c *Client) AttachContainer(id string, q *AttachContainerRequest) (io.ReadWriteCloser, error) {
  err := validateId(id)
  if err != nil {
    return nil, err
  }
  res, client, err := c.callfquery2("POST","/v1.4/containers/%s/attach",q.Values(), id)
  if err != nil {
    if client != nil {
      client.Close()
    }
    return nil, err
  }
  if err = expectHTTPStatus(res.StatusCode, 200); err != nil {
    return nil, err
  }
  con, buf := client.Hijack()
  return &hijackReadWriteCloser{con,buf}, nil
}

