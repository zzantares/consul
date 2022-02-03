@setupApplicationTest
Feature: dc / services / instances / health-checks
  Background:
    Given 1 datacenter model with the value "dc1"
  Scenario: 2 Service Instances only one having a connect proxy
    Given 1 proxy model from yaml
    ---
    - ServiceName: service-0-sidecar-proxy
      ServiceID: service-0-b-sidecar-proxy
      ServiceProxy:
        DestinationServiceName: service-0
        DestinationServiceID: service-0-b
      Node: node-0
    ---
    And 2 instance models from yaml
    ---
    - Service:
        Name: service-0-sidecar-proxy
        ID: service-0-b-sidecar-proxy
      Node:
        Node: node-0
      Checks:
        - Type: 'tcp'
          Name: Connect Sidecar Listening
          CheckID: "service:service-0-sidecar-proxy:1"
          ServiceID: service-0-b-sidecar-proxy
          ServiceName: service-0-sidecar-proxy
          Status: critical
          Output: No checks found.
    - Service:
        Name: service-0
        ID: service-0-a
      Node:
        Node: node-0
      Checks:
        - Type: ''
          Name: Serf Health Status
          CheckID: serfHealth
          Status: critical
          Output: ouch
    ---
    When I visit the instance page for yaml
    ---
      dc: dc1
      service: service-0
      node: node-0
      id: service-0-a
    ---
    Then the url should be /dc1/services/service-0/instances/node-0/service-0-a/health-checks
    And I see healthChecksIsSelected on the tabs
    And I don't see the "[data-test-health-check-id='service:service-0-sidecar-proxy:1']" element

  Scenario: A failing serf check
    Given 1 proxy model from yaml
    ---
    - ServiceProxy:
        DestinationServiceName: service-1
        DestinationServiceID: ~
    ---
    And 2 instance models from yaml
    ---
    - Service:
        ID: service-0-with-id
      Node:
        Node: node-0
    - Service:
        ID: service-1-with-id
      Node:
        Node: another-node
      Checks:
        - Type: ''
          Name: Serf Health Status
          CheckID: serfHealth
          Status: critical
          Output: ouch
    ---
    When I visit the instance page for yaml
    ---
      dc: dc1
      service: service-0
      node: another-node
      id: service-1-with-id
    ---
    Then the url should be /dc1/services/service-0/instances/another-node/service-1-with-id/health-checks
    And I see healthChecksIsSelected on the tabs
    And I see criticalSerfNotice on the tabs.healthChecksTab
  Scenario: A passing serf check
    Given 1 proxy model from yaml
    ---
    - ServiceProxy:
        DestinationServiceName: service-1
        DestinationServiceID: ~
    ---
    And 2 instance models from yaml
    ---
    - Service:
        ID: service-0-with-id
      Node:
        Node: node-0
    - Service:
        ID: service-1-with-id
      Node:
        Node: another-node
      Checks:
        - Type: ''
          Name: Serf Health Status
          CheckID: serfHealth
          Status: passing
          Output: Agent alive and reachable
    ---
    When I visit the instance page for yaml
    ---
      dc: dc1
      service: service-0
      node: another-node
      id: service-1-with-id
    ---
    Then the url should be /dc1/services/service-0/instances/another-node/service-1-with-id/health-checks
    And I see healthChecksIsSelected on the tabs
    And I don't see criticalSerfNotice on the tabs.healthChecksTab
