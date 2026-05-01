Feature: AI PDF risk summarization

  @BEHAVIOR-001
  Scenario: Ambiguous risk clause is flagged
    Given a financial PDF says exposure "may be material subject to market conditions"
    When the system summarizes the risk clause
    Then the summary should flag the exposure as uncertain

  @BEHAVIOR-001
  Scenario: Ambiguous risk clause is not stated as fact
    Given a financial PDF uses qualified risk language
    When the system produces a risk summary
    Then the summary should not state the risk as confirmed exposure
