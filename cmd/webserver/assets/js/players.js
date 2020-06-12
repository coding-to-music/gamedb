if ($('#players-page').length > 0) {

    const countriesWithStates = ['AU', 'CA', 'FR', 'GB', 'NZ', 'PH', 'SI', 'US'];

    const $country = $('#country');
    const $search = $('#search');

    $country.on('change', function (e) {
        toggleStateDropDown();
    });

    function toggleStateDropDown() {

        const countryVal = $country.val();

        for (const cc of countriesWithStates) {
            if (cc === countryVal) {
                $('.state-dd[data-cc="' + cc + '"]').removeClass('d-none');
            } else {
                $('.state-dd[data-cc="' + cc + '"]').addClass('d-none');
            }
        }
    }

    const options = {
        "language": {
            "zeroRecords": function () {
                return 'Players can be searched for using their username or vanity URL. If a player is missing, <a href="/players/add?search=' + $search.val() + '">add them here</a>.';
            },
        },
        "order": [[3, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[13]);
        },
        "columnDefs": [
            // Rank
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[0];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('font-weight-bold')
                },
                "orderable": false,
            },
            // Flag
            {
                "targets": 1,
                "render": function (data, type, row) {
                    if (row[11]) {
                        const img = '<img data-lazy="' + row[11] + '" alt="" data-lazy-alt="' + row[12] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[12] + '">';
                        return '<a href="/players?country=' + row[19] + '">' + img + '</a>';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Player
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + encodeHTML(row[2]) + '"></div><div class="name">' + encodeHTML(row[2]) + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                },
                "orderable": false,
            },
            // Avatar 2 / Level
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><div class="' + row[4] + '"></div></div><div class="name min nowrap">' + row[5].toLocaleString() + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderSequence": ["desc"],
            },
            // Badges
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[7].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Games
            {
                "targets": 5,
                "render": function (data, type, row) {

                    if (row[6]) {
                        return row[6].toLocaleString();
                    }
                    return $lockIcon;
                },
                "orderSequence": ["desc"],
            },
            // Time
            {
                "targets": 6,
                "render": function (data, type, row) {

                    if (row[8] === '-') {
                        return $lockIcon;
                    }

                    return row[8];
                },
                "createdCell": function (td, cellData, rowData, row, col) {

                    if (rowData[8] !== '-') {
                        const $td = $(td);
                        $td.attr('nowrap', 'nowrap');
                        $td.attr('data-toggle', 'tooltip');
                        $td.attr('data-placement', 'left');
                        $td.attr('title', rowData[9]);
                    }
                },
                "orderSequence": ["desc"],
            },
            // Game Bans
            {
                "targets": 7,
                "render": function (data, type, row) {
                    return row[15].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // VAC Bans
            {
                "targets": 8,
                "render": function (data, type, row) {
                    return row[16].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Last Ban
            {
                "targets": 9,
                "render": function (data, type, row) {
                    if (row[17] > 0) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[18] + '" data-livestamp="' + row[17] + '"></span>';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderSequence": ["desc"],
            },
            // Friends
            {
                "targets": 10,
                "render": function (data, type, row) {

                    if (row[10] === 0) {
                        return $lockIcon;
                    }

                    return row[10].toLocaleString();
                },
                "orderSequence": ["desc"],
                "visible": false,
            },
            // Comments
            {
                "targets": 11,
                "render": function (data, type, row) {
                    return row[20].toLocaleString();
                },
                "orderSequence": ["desc"],
                "visible": false,
            },
            // Link
            {
                "targets": 12,
                "render": function (data, type, row) {
                    if (row[14]) {
                        return '<a href="' + row[14] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };

    let searchFields = [
        $country,
        $search,
    ];

    for (const cc of countriesWithStates) {
        searchFields.push($('#' + cc + '-state'));
    }

    const dt = $('table.table').gdbTable({tableOptions: options, searchFields: searchFields});

    function updateColumns(dt, hash) {

        if (!hash) {
            hash = '#level';
        }

        $('#player-nav a[href="' + hash + '"]').tab('show');

        const oldOrder = dt.order();

        let hide = [];
        let show = [];

        switch (hash) {
            case '#level':

                show = [3, 4];
                hide = [5, 6, 7, 8, 9, 10, 11];

                dt.order([3, 'desc']);
                break;

            case '#games':

                show = [5, 6];
                hide = [3, 4, 7, 8, 9, 10, 11];

                dt.order([5, 'desc']);
                break;

            case '#bans':

                show = [7, 8, 9];
                hide = [3, 4, 5, 6, 7, 8, 9, 10, 11];

                dt.order([7, 'desc']);
                break;

            case '#profile':

                show = [10, 11];
                hide = [3, 4, 5, 6, 7, 8, 9, 10, 11];

                dt.order([10, 'desc']);
                break;
        }

        hide.forEach(function (value, index, array) {
            dt.column(value).visible(false);
        });

        show.forEach(function (value, index, array) {
            dt.column(value).visible(true);
        });

        if (JSON.stringify(oldOrder) !== JSON.stringify(dt.order())) {
            dt.draw();
        }

        const table = dt.table().container();
        observeLazyImages($(table).find('img[data-lazy]'));
    }

    $('#player-nav a[href^="#"]').on('click', function (e) {

        e.preventDefault();

        const href = $(this).attr('href');

        window.location.hash = href;
        updateColumns(dt, href);
    });

    setTimeout(
        function () {
            updateColumns(dt, window.location.hash);
        },
        1000
    );

    toggleStateDropDown();
}
