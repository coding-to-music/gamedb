if ($('#players-page').length > 0) {

    const $country = $('#country');

    $country.on('change', function (e) {
        toggleStateDropDown();
    });

    function toggleStateDropDown() {

        const $container = $('#state-container');
        if ($country.val() === 'US') {
            $container.removeClass('d-none');
        } else {
            $container.addClass('d-none');
        }
    }

    toggleStateDropDown();

    const options = {
        "language": {
            "zeroRecords": "No players found <a href='/players/add'>Add a Player</a>",
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
                        return '<img data-lazy="' + row[11] + '" alt="" data-lazy-alt="' + row[12] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[12] + '">';
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
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" alt="" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
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
                    return '<div class="icon-name"><div class="icon"><div class="' + row[4] + '"></div></div><div class="name min">' + row[5].toLocaleString() + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderSequence": ["desc"],
            },
            // Games
            {
                "targets": 4,
                "render": function (data, type, row) {

                    if (row[6]) {
                        return row[6].toLocaleString();
                    }
                    return $lockIcon;
                },
                "orderSequence": ["desc"],
            },
            // Badges
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return row[7].toLocaleString();
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

                    $(td).attr('nowrap', 'nowrap');

                    if (rowData[8] !== '0m') {
                        $(td).attr('data-toggle', 'tooltip').attr('data-placement', 'left').attr('title', rowData[9]);
                    }
                },
                "orderSequence": ["desc"],
            },
            // Friends
            {
                "targets": 7,
                "render": function (data, type, row) {

                    if (row[10] === 0) {
                        return $lockIcon;
                    }

                    return row[10].toLocaleString();
                },
                "orderSequence": ["desc"],
            },

            // Game Bans
            {
                "targets": 8,
                "render": function (data, type, row) {
                    return row[15].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // VAC Bans
            {
                "targets": 9,
                "render": function (data, type, row) {
                    return row[16].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Last Ban
            {
                "targets": 10,
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

            // Link
            {
                "targets": 11,
                "render": function (data, type, row) {
                    if (row[14]) {
                        return '<a href="' + row[14] + '" target="_blank" rel="nofollow"><i class="fas fa-link" data-target="_blank"></i></a>';
                    }
                    return '';
                },
                "orderable": false,
            },
        ]
    };

    const searchFields = [
        $('#search'),
        $('#state'),
        $country,
    ];

    const dt = $('table.table').gdbTable({tableOptions: options, searchFields: searchFields});

    function updateColumns(dt, hash) {

        if (!hash) {
            hash = '#stats';
        }

        $('#player-nav a[href="' + hash + '"]').tab('show');

        switch (hash) {
            case '#stats':

                dt.column(3).visible(true);
                dt.column(4).visible(true);
                dt.column(5).visible(true);
                dt.column(6).visible(true);
                dt.column(7).visible(true);

                dt.column(8).visible(false);
                dt.column(9).visible(false);
                dt.column(10).visible(false);
                break;

            case '#bans':

                dt.column(3).visible(false);
                dt.column(4).visible(false);
                dt.column(5).visible(false);
                dt.column(6).visible(false);
                dt.column(7).visible(false);

                dt.column(8).visible(true);
                dt.column(9).visible(true);
                dt.column(10).visible(true);
                break;
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
}
