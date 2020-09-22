const $settingsPage = $('#settings-page');

if ($settingsPage.length > 0 || $('#signup-page').length > 0) {

    $('#password-container input:password').pwstrength();
}

if ($settingsPage.length > 0) {

    // Browser alert permissions
    // const $checkbox = $('#browser-alerts');
    //
    // $checkbox.on('click', function () {
    //     if ($(this).is(':checked')) {
    //
    //         Push.Permission.request(
    //             function () {
    //             },
    //             function () {
    //                 alert('You have denied notification access in your browser.');
    //                 $(this).prop("checked", false);
    //             }
    //         );
    //     }
    // });

    // On tab change
    $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

        const to = $(e.target);
        const from = $(e.relatedTarget);

        // On entering tab
        if (!to.attr('loaded')) {
            to.attr('loaded', 1);
            switch (to.attr('href')) {
                case '#events':
                    loadEvents();
                    break;
                case '#donations':
                    loadDonations();
                    break;
            }
        }
    });

    function loadEvents() {

        const options = {
            "order": [[0, 'desc']],
            "columnDefs": [
                // Time
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[1] + '" data-livestamp="' + row[0] + '">' + row[1] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
                // Type
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<i class="fas ' + row[7] + '"></i> ' + row[2];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
                // IP
                {
                    "targets": 2,
                    "render": function (data, type, row) {

                        if (row[3] === row[6]) {
                            return '<span class="font-weight-bold" data-toggle="tooltip" data-placement="left" title="Your current IP">' + row[3] + '</span>';
                        }
                        return row[3];
                    },
                    "orderable": false
                },
                // User Agent
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        // return row[4];
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '">' + row[5] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                }
            ]
        };

        $('#events table.table').gdbTable({tableOptions: options});
    }

    function loadDonations() {

        const options = {
            "order": [[0, 'desc']],
            "columnDefs": [
                // Time
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return '<span data-toggle="tooltip" data-placement="left" title="' + row[1] + '" data-livestamp="' + row[0] + '">' + row[1] + '</span>';
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
                // Type
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<i class="fas ' + row[7] + '"></i> ' + row[2];
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).attr('nowrap', 'nowrap');
                    },
                    "orderable": false
                },
            ]
        };

        $('#donations table.table').gdbTable({tableOptions: options});
    }
}
