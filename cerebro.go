package main

import (
  "fmt"
  "strings"
  "log"
  "time"
  "os"
  //"net/http"
  "github.com/gin-gonic/gin"
  "github.com/samalba/dockerclient"
  "github.com/coreos/go-etcd/etcd"
)

type dockerManagedInstance struct {
  dockerId string
  appName string
  version int
  instanceType string
  instanceNumber int
  instanceIp string
  instancePort int
}

// FIXME

func KnownAppNames() []string {
  return []string{"pga-staging", "pga-production"}
}

func ActiveReleaseVersion() int {
  return 1
}

// DOCKER

func DockerClient() *dockerclient.DockerClient {
  docker, _ := dockerclient.NewDockerClient(os.Getenv("DOCKER_ENDPOINT"), nil)
  return docker
}

func DockerHandler(instanceUp chan<- dockerManagedInstance) {
  dockerClient := DockerClient()

  ticker := time.NewTicker(time.Millisecond * 1000)
  go func() {
    for _ = range ticker.C {
      DockerScanAllInstances(dockerClient, instanceUp)
    }
  }()

  // FIXME: How do pass arguments?
  //dockerClient.StartMonitorEvents(DockerEventCallback, nil)
}

func DockerEventCallback(event *dockerclient.Event, ec chan error, instanceUp chan<- dockerManagedInstance, args ...interface{}) {
  log.Printf("Received event: %#v\n", *event)

  switch event.Status {
  case "start":
    if found, instance := DockerIdentify(event.Id); found {
      instanceUp <- instance
    }
  case "die":
    // Note: We don't handle this for now since we have our TTL to expire dead containers
  }
}

func DockerScanAllInstances(dockerClient *dockerclient.DockerClient, instanceUp chan<- dockerManagedInstance) {
  log.Printf("Scanning all instances...")

  containers, err := dockerClient.ListContainers(true, false, "")

  if err != nil {
    log.Print("Failed to list docker containers")
    return
  }

  for _, container := range containers {
    if found, instance := DockerIdentify(container.Id); found {
      instanceUp <- instance
    }
  }
}

func DockerIdentify(dockerId string) (bool, dockerManagedInstance) {
  instance := dockerManagedInstance{dockerId: dockerId}

  containerInfo, err := DockerClient().InspectContainer(dockerId)
  if err != nil {
    log.Printf("%#v", err)
    return false, instance
  }

  for _, appName := range KnownAppNames() {
    if strings.HasPrefix(containerInfo.Name[1:], appName) {
      instance.appName = appName
      break
    }
  }

  if instance.appName == "" {
    return false, instance
  }

  parts := strings.Split(containerInfo.Name, "-")
  if len(parts) < 4 {
    return false, instance
  }

  fmt.Sscanf(parts[len(parts)-1], "%d",  &instance.instanceNumber)
  fmt.Sscanf(parts[len(parts)-2], "v%d", &instance.version)
  instance.instanceType = parts[len(parts)-3]

  if instance.instanceNumber == 0 || instance.instanceType == "" {
    return false, instance
  }

  if instance.version != ActiveReleaseVersion() {
    log.Printf("Found container with invalid release v%d", instance.version)
    return false, instance
  }

  instance.instanceIp = containerInfo.NetworkSettings.IpAddress

  for port := range containerInfo.NetworkSettings.Ports {
    fmt.Sscanf(strings.Split(port, "/")[0], "%d", &instance.instancePort)
    break
  }

  return true, instance
}

// ETCD STATE

func EtcdClient() *etcd.Client {
  etcdMachines := []string{os.Getenv("ETCD_ENDPOINT")}
  return etcd.NewClient(etcdMachines)
}

func EtcdKeyForInstance(instance dockerManagedInstance) string {
  return fmt.Sprintf("/cerebro/%v/releases/%v/instances/%v/%v",
                     instance.appName, instance.version,
                     instance.instanceType, instance.instanceNumber)
}

func InstanceUp(etcdClient *etcd.Client, instance dockerManagedInstance) {
  log.Printf("%#v", instance)

  ipAndPort := fmt.Sprintf("%v:%v", instance.instanceIp, instance.instancePort)
  etcdKey   := EtcdKeyForInstance(instance)

  if _, err := etcdClient.Set(etcdKey, ipAndPort, 10); err != nil {
    log.Print(err)
  }
}

func ListenForInstanceUp(instanceUp <-chan dockerManagedInstance) {
  etcdClient := EtcdClient()
  for instance := range instanceUp {
    InstanceUp(etcdClient, instance)
  }
}

// OTHER
func HandleAllInstancesUp() {
  // - Switches active load balancer config version
  // - SET /cerebro/APPNAME/release to VERSION
  // - stop all old instances (using fleet)
  // - (future) Calls WEBHOOK
  // - (future) Prunes oldest release (so release log has a total of 5)
}

func HealthCheck(instance dockerManagedInstance) {
  // Check that container is running
  // Check that container return a 200 result for CHECK_URL at PORT
  // Return true/false
}

func DeployRelease() {
  // - Start new instances (through fleet)
  // - initialize new release
}

func CheckForNewRelease() {
  // - Every cerebro instance: Waits for all instances to be up and registered under the new version
  // - Acquire etcd lock (exit routine if not acquired)
  HandleAllInstancesUp()
}

func main() {
  router := gin.Default()

  /*router.GET("/", func(c *gin.Context) {
    containers, err := DockerClient().ListContainers(true, false, "")
    if err != nil {
      log.Printf("Error: %#v\n", err)
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not access container list"})
    }
    for _, container := range containers {
      c.JSON(http.StatusOK, gin.H{"id": container.Id, "names": container.Names})
    }
  })*/

  instanceUp := make(chan dockerManagedInstance)

  go ListenForInstanceUp(instanceUp)

  go DockerHandler(instanceUp)

  router.Run(":" + os.Getenv("PORT"))
}
