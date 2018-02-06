@statusapi
Feature: Model state for events
  Scenario:
    Given a milestone model
    And some correlated events for the model
    When I retrieve the model state for the correlated events
    Then the state of the model reflects the events