if ($('#developers-page').length > 0) {

    var options = {
        valueNames: ['developer-name'],
        listClass: 'developers-list',
        page: 1000,
        fuzzySearch: {
            searchClass: 'developers-search',
            location: 0,
            threshold: 0.5
        }
    };

    new List('developers-page', options);
}
