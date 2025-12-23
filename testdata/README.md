# How to Testing with the test data

```bash
# Scenario 1 - All matched
./bin/recon -system=testdata/scenario1_all_matched_system.csv -banks=testdata/scenario1_all_matched_bank_bca.csv,testdata/scenario1_all_matched_bank_bri.csv -start=2024-01-15 -end=2024-01-30

# Scenario 2 - Unmatched system
./bin/recon -system=testdata/scenario2_system_unmatched_system.csv -banks=testdata/scenario2_system_unmatched_bank_bca.csv -start=2024-01-01 -end=2024-01-31

# Scenario 3 - Unmatched bank
./bin/recon -system=testdata/scenario3_bank_unmatched_system.csv -banks=testdata/scenario3_bank_unmatched_bank_bca.csv -start=2024-01-01 -end=2024-01-31

# Scenario 4 - Both unmatched
./bin/recon -system=testdata/scenario4_both_unmatched_system.csv -banks=testdata/scenario4_both_unmatched_bank_bca.csv,testdata/scenario4_both_unmatched_bank_mandiri.csv -start=2024-01-01 -end=2024-01-31
```
