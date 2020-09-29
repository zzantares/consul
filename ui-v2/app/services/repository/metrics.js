import RepositoryService from 'consul-ui/services/repository';
export default RepositoryService.extend({
  findByService: function(slug, dc, nspace, configuration = {}) {
    const result = {
      items: [Math.round(Math.random() * 10)],
    };
    // cursor can be anything
    // but it must be on result.meta.cursor
    result.meta = {
      cursor: 1,
    };
    return Promise.resolve(result);
  },
});
