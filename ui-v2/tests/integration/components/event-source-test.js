import { moduleForComponent } from 'ember-qunit';
import test from 'ember-sinon-qunit/test-support/test';
import hbs from 'htmlbars-inline-precompile';

moduleForComponent('event-source', 'Integration | Component | event source', {
  integration: true,
});
const $ = function(sel) {
  return document.querySelectorAll(sel);
};
test('it dispatches events to its previous sibling when no ref is defined', function(assert) {
  this.render(hbs`
    {{#event-source as |src target dispatch|}}
      <button class="button" {{ action dispatch 'remove'}}>click to remove</button>
      {{event-source src=src onremove=(action target 'remove')}}
    {{/event-source}}
  `);
  assert.equal(
    this.$()
      .text()
      .trim(),
    'click to remove'
  );
  $('.button')[0].click();
  assert.equal(
    this.$()
      .text()
      .trim(),
    ''
  );
});
test('it dispatches events to its ref when ref is defined', function(assert) {
  // Set any properties with this.set('myProperty', 'value');
  const focus = this.stub();
  this.set('ref', {
    focus: focus,
  });
  this.render(hbs`
    {{#event-source as |src target dispatch|}}
      <button class="button" {{ action dispatch 'focus'}}>click to focus</button>
      {{event-source src=src ref=ref onfocus=(action target 'focus')}}
    {{/event-source}}
  `);
  assert.equal(
    this.$()
      .text()
      .trim(),
    'click to focus'
  );
  $('.button')[0].click();
  assert.equal(
    this.$()
      .text()
      .trim(),
    'click to focus'
  );
  assert.ok(focus.calledOnce);
});
