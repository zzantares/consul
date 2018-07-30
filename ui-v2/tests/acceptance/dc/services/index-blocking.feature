@setupApplicationTest
Feature: dc / services / list
  In order to see updates without refreshing the page
  As a user
  I want to see a newly registered services as I add them via other means
  Scenario:
    Given 1 datacenter model with the value "dc-1"
    And settings from yaml
    ---
    autorefresh: 1
    ---
    And 3 service models
    And a network latency of 100
    When I visit the services page for yaml
    ---
      dc: dc-1
    ---
    Then the url should be /dc-1/services
    And pause until I see 3 service models
    And an external edit results in 5 service models
    And pause until I see 5 services models
