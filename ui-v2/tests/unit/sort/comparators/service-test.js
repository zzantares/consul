import comparatorFactory from 'consul-ui/sort/comparators/service';
import { module, test } from 'qunit';

module('Unit | Sort | Comparator | service', function() {
  const comparator = comparatorFactory();
  test('Passing anything but Status: just returns what you gave it', function(assert) {
    const expected = 'Name:asc';
    const actual = comparator(expected);
    assert.equal(actual, expected);
  });
  test('items are sorted by a fake Status which uses PercentageMeshChecks{Passing,Warning,Critical}', function(assert) {
    const items = [
      {
        PercentageMeshChecksPassing: 100,
        PercentageMeshChecksWarning: 0,
        PercentageMeshChecksCritical: 0,
      },
      {
        PercentageMeshChecksPassing: 50,
        PercentageMeshChecksWarning: 20,
        PercentageMeshChecksCritical: 30,
      },
      {
        PercentageMeshChecksPassing: 50,
        PercentageMeshChecksWarning: 40,
        PercentageMeshChecksCritical: 10,
      },
      {
        PercentageMeshChecksPassing: 20,
        PercentageMeshChecksWarning: 10,
        PercentageMeshChecksCritical: 70,
      },
    ];
    const comp = comparator('Status:desc');
    assert.equal(typeof comp, 'function');

    let actual = items.sort(comp);
    assert.deepEqual(actual, items);

    const expected = [...items];
    expected.reverse();
    actual = items.sort(comparator('Status:asc'));
    assert.deepEqual(actual, expected);
  });
});
