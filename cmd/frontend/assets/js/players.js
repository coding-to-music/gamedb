if ($('#players-page').length > 0) {

    const $search = $('#search');
    const $country = $('#country');
    const $state = $('#state');
    const $stateContainer = $('#state-container');

    $country.on('change', function (e) {
        toggleStateDropDown();
    });

    function toggleStateDropDown() {

        const countryVal = $country.val();
        const isContient = countryVal.startsWith("c-");

        if (isContient) {
            $stateContainer.hide();
            return;
        }

        $.ajax({
            type: "GET",
            url: '/players/states.json?cc=' + countryVal,
            dataType: 'json',
            success: function (data, textStatus, jqXHR) {

                if (!data || data.length <= 3) {
                    $stateContainer.hide();
                    return;
                }

                $state.empty();
                $.each(data, function (index, value) {

                    let ops = {
                        value: value['k'],
                        text: value['v'],
                    };

                    if (value['v'] === '---') {
                        ops['disabled'] = 'disabled';
                    }

                    $state.append($('<option/>', ops));
                });

                $state.val('');
                $stateContainer.show();
            },
        });
    }

    const options = {
        "language": {
            "zeroRecords": function () {
                return 'Players can be searched for using their username or vanity URL. If a player is missing, <a href="/players/add?search=' + $search.val() + '">add them here</a>.';
            },
        },
        "order": [getOrder(window.location.hash)],
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
                    return '<a href="' + row[13] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[23] + '</div></a>'
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
            // Achievements
            {
                "targets": 12,
                "render": function (data, type, row) {
                    return row[21].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Achievements
            {
                "targets": 13,
                "render": function (data, type, row) {
                    return row[22].toLocaleString();
                },
                "orderSequence": ["desc"],
            },

            // Foil Badges
            {
                "targets": 14,
                "render": function (data, type, row) {
                    return row[25].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Link
            {
                "targets": 15,
                "render": function (data, type, row) {
                    if (row[14]) {
                        return '<a href="' + row[14] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
            // Search Score
            {
                "targets": 16,
                "render": function (data, type, row) {
                    return row[24];
                },
                "orderable": false,
                "visible": false,
            },
        ]
    };

    let searchFields = [
        $country,
        $search,
        $state,
    ];

    const dt = $('table.table').gdbTable({tableOptions: options, searchFields: searchFields});

    function getOrder(hash) {

        const searchSortCol = $('[data-col-sort]').attr('data-col-sort');

        if ($search.val() && searchSortCol) {
            return [parseInt(searchSortCol), 'desc'];
        }

        switch (hash) {
            case '#games':
                return [5, 'desc'];
            case '#bans':
                return [7, 'desc'];
            case '#profile':
                return [10, 'desc'];
            case '#achievements':
                return [12, 'desc'];
            default:
                // Level
                return [3, 'desc'];
        }
    }

    function updateColumns(dt, hash) {

        if (!hash) {
            hash = '#level';
        }

        // Update tab UI
        $('#player-nav a[href="' + hash + '"]').tab('show');

        let hide = [];
        let show = [];

        switch (hash) {
            case '#level':
                show = [3, 4, 14];
                hide = [5, 6, 7, 8, 9, 10, 11, 12, 13];
                break;
            case '#games':
                show = [5, 6];
                hide = [3, 4, 7, 8, 9, 10, 11, 12, 13, 14];
                break;
            case '#bans':
                show = [7, 8, 9];
                hide = [3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14];
                break;
            case '#profile':
                show = [10, 11];
                hide = [3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14];
                break;
            case '#achievements':
                show = [12, 13];
                hide = [3, 4, 5, 6, 7, 8, 9, 10, 11, 14];
                break;
        }

        dt.columns(hide).visible(false, false);
        dt.columns(show).visible(true, false);
        dt.columns.adjust(); // Adjust column sizing

        if (JSON.stringify(dt.order()) !== JSON.stringify([getOrder(hash)])) {
            dt.order([getOrder(hash)]);
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

    updateColumns(dt, window.location.hash);
    toggleStateDropDown();
}
