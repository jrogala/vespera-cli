Feature: List observations
  As a user I can list all observations stored on the telescope.

  Background:
    Given a connected telescope

  Scenario: List observations returns entries
    Given the telescope has observations:
      | name          | date             |
      | M42-Orion     | 2024-01-15 21:30 |
      | NGC7000       | 2024-01-16 22:00 |
    When I list observations
    Then I should get 2 observations
    And observation "M42-Orion" should be in the list
    And observation "NGC7000" should be in the list

  Scenario: List observations when empty
    Given the telescope has 0 observations
    When I list observations
    Then I should get 0 observations
