@setupApplicationTest
Feature: dc / services / list
  In order to see registered services
  As a user
  I want to see a list of registered services when I visit the service index page
  Scenario:
    Given 1 datacenter model with the value "dc-1"
    And 3 service models
    When I visit the services page for yaml
    ---
      dc: dc-1
    ---
    Then the url should be /dc-1/services
    Then I see 3 service models
