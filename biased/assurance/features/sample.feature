Feature: Sample system validation

  Scenario: System behaves as intended
    Given the system is initialized
    When validation artifacts are present
    Then the outcomes should satisfy BIASED
