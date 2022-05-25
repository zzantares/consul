(routes => routes({
  dc: {
    peers: {
      _options: {
        path: '/peers',
        abilities: ['read peers'],
      },
      index: {
        _options: {
          path: '/',
          queryParams: {
            sortBy: 'sort',
            searchproperty: {
              as: 'searchproperty',
              empty: [['Name', 'Description']],
            },
            search: {
              as: 'filter',
              replace: true,
            },
          },
        },
      },
    },
  },
}))(
  (json, data = (typeof document !== 'undefined' ? document.currentScript.dataset : module.exports)) => {
    data[`routes`] = JSON.stringify(json);
  }
);
