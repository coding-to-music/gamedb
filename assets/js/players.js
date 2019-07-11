if ($('#ranks-page').length > 0) {

    const $playersTable = $('table.table-datatable2');

    $('form').on('submit', function (e) {

        $playersTable.DataTable().draw();
        return false;
    });

    $('#country, #state').on('change', function (e) {

        $playersTable.DataTable().draw();
        toggleStateDropDown();
        return false;
    });

    function toggleStateDropDown() {

        $container = $('#state-container');
        if ($('#country').val() === 'US') {
            $container.removeClass('d-none');
        } else {
            $container.addClass('d-none');
        }
    }

    toggleStateDropDown();

    $playersTable.DataTable($.extend(true, {}, dtDefaultOptions, {
        "ajax": function (data, callback, settings) {

            data.search = {};
            data.search.search = $('#search').val();
            data.search.country = $('#country').val();
            data.search.state = $('#state').val();

            dtDefaultOptions.ajax(data, callback, settings, $(this));
        },
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
            // Player
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                },
                "orderable": false,
            },
            // Flag
            {
                "targets": 2,
                "render": function (data, type, row) {
                    if (row[11]) {
                        return '<img data-lazy="' + row[11] + '" data-lazy-alt="' + row[12] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[12] + '">';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
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
            }
        ]
    }));
}