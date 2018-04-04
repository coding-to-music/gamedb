if ($('#publishers-page').length > 0) {

    var options = {
        valueNames: ['publisher-name'],
        listClass: 'publishers-list',
        page: 1000,
        fuzzySearch: {
            searchClass: 'publishers-search',
            location: 0,
            threshold: 0.5
        }
    };

    new List('publishers-page', options);
}
