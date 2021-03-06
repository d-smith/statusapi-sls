@statusapi
Feature: Model state for events
  @model
  Scenario:
    Given a milestone model
    And some correlated events for the model
    When I retrieve the model state for the correlated events
    Then the state of the model reflects the events

  @modelupdate
  Scenario:
    Given a milestone model and correlated events
    When I update the model
    Then the model state reflects the update
    And the model update is durable
    And the model reflects future events

  @events
  Scenario:
    When I post events for a transaction
    Then I can retrieve those events using the transaction id