if ($('#tags-page').length > 0) {

    var options = {
        valueNames: ['tag-name'],
        listClass: 'tags-list',
        page: 1000,
        fuzzySearch: {
            searchClass: 'tags-search',
            location: 0,
            threshold: 0.5
        }
    };

    new List('tags-page', options);
}
