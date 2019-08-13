if ($('#player-bans-page').length > 0) {

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
            $(row).attr('data-link', data[6]);
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
                    if (row[4]) {
                        return '<img data-lazy="' + row[4] + '" data-lazy-alt="' + row[5] + '" class="wide" data-toggle="tooltip" data-placement="left" data-lazy-title="' + row[5] + '">';
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
                    return '<div class="icon-name"><div class="icon"><img data-lazy="' + row[3] + '" data-lazy-alt="' + row[2] + '"></div><div class="name">' + row[2] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img')
                },
                "orderable": false,
            },
            // Game Bans
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[7].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // VAC Bans
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return row[8].toLocaleString();
                },
                "orderSequence": ["desc"],
            },
            // Last Ban
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderSequence": ["desc"],
            },
        ]
    }));
}