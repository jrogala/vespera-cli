Feature: Telescope status
  As a user I can check if the telescope is reachable and see how many observations exist.

  Background:
    Given a connected telescope

  Scenario: Status shows host and observation count
    Given the telescope has 3 observations
    When I request the telescope status
    Then I should see the host address
    And I should see 3 observations

  Scenario: Status with no observations
    Given the telescope has 0 observations
    When I request the telescope status
    Then I should see the host address
    And I should see 0 observations
