const $settingsPage = $('#settings-page');

if ($settingsPage.length > 0 || $('#signup-page').length > 0) {

    $('#password-container input:password').pwstrength();
}

if ($settingsPage.length > 0) {

    $('#reset-api-key-btn').on('click', function (e) {

        return bootbox.confirm({
            message: "Are you sure? Any calls using the current API key will stop wirking.",
            buttons: {
                confirm: {
                    label: 'Yes',
                    className: 'btn-success'
                },
                cancel: {
                    label: 'No',
                    className: 'btn-danger'
                }
            },
            callback: function (result) {
                if (result) {
                    window.location.replace('/settings/new-key');
                }
            }
        });
    });

    loadAjaxOnObserve({
        'events-table': loadEvents,
        'donations-table': loadDonations,
    });

    function loadEvents() {

        // Setup drop downs
        $('select.form-control-chosen').chosen({
            disable_search_threshold: 10,
            allow_single_deselect: true,
            rtl: false,
        });

        //
        $('#events table.table').gdbTable({
            searchFields: [
                $('#type'),
            ],
            tableOptions: {
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
                    // Event Type
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
                    // Location (IP)
                    {
                        "targets": 2,
                        "render": function (data, type, row) {

                            if (row[3] === row[6]) {
                                return '<span class="font-weight-bold" data-toggle="tooltip" data-placement="left" title="Your current IP">' + row[8] + '</span>';
                            }
                            return row[8];
                        },
                        "orderable": false
                    },
                    // User Agent
                    {
                        "targets": 3,
                        "render": function (data, type, row) {
                            return '<span data-toggle="tooltip" data-placement="left" title="' + row[4] + '">' + row[5] + '</span>';
                        },
                        "createdCell": function (td, cellData, rowData, row, col) {
                            $(td).attr('nowrap', 'nowrap');
                        },
                        "orderable": false
                    }
                ]
            }
        });
    }

    function loadDonations() {

        $('#donations table.table').gdbTable({
            tableOptions: {
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
                            return '$ ' + (row[2] / 100).toLocaleString();
                        },
                        "createdCell": function (td, cellData, rowData, row, col) {
                            $(td).attr('nowrap', 'nowrap');
                        },
                        "orderable": false
                    },
                    // Source
                    {
                        "targets": 2,
                        "render": function (data, type, row) {
                            return row[3];
                        },
                        "orderable": false
                    },
                ]
            }
        });
    }
}
