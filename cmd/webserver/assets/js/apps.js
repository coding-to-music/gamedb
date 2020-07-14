if ($('#apps-page').length > 0) {

    (function ($, document) {
        'use strict';

        // Setup drop downs
        $('select.form-control-chosen').chosen({
            disable_search_threshold: 10,
            allow_single_deselect: true,
            rtl: false,
            max_selected_options: 10
        });

        // Price slider
        const $priceElement = $('#price-slider');
        const priceSlider = noUiSlider.create($priceElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100,
            }
        });

        // Score slider
        const $scoreElement = $('#score-slider');
        const scoreSlider = noUiSlider.create($scoreElement[0], {
            start: [0, 100],
            connect: true,
            step: 1,
            range: {
                'min': 0,
                'max': 100
            }
        });

        //
        function updateLabels(e) {

            const prices = priceSlider.get();
            if (prices[0] === prices[1]) {
                $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ')');
            } else {
                $('label#price-label').html('Price (' + user.userCurrencySymbol + Math.round(prices[0]) + ' - ' + user.userCurrencySymbol + Math.round(prices[1]) + ')');
            }

            const scores = scoreSlider.get();
            if (scores[0] === scores[1]) {
                $('label#score-label').html('Score (' + Math.round(scores[0]) + '%)');
            } else {
                $('label#score-label').html('Score (' + Math.round(scores[0]) + '% - ' + Math.round(scores[1]) + '%)');
            }
        }

        window.updateLabels = updateLabels;

        $(updateLabels);

        // Setup datatable
        const options = {
            "order": [[2, 'desc']],
            "createdRow": function (row, data, dataIndex) {
                $(row).attr('data-app-id', data[0]);
                $(row).attr('data-link', data[3]);
            },
            "columnDefs": [
                // Rank
                {
                    "targets": 0,
                    "render": function (data, type, row) {
                        return row[9].toLocaleString();
                    },
                    "orderable": false,
                },
                // Icon / App Name
                {
                    "targets": 1,
                    "render": function (data, type, row) {
                        return '<a href="' + row[3] + '" class="icon-name"><div class="icon"><img data-lazy="' + row[2] + '" alt="" data-lazy-alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></a>'
                    },
                    "createdCell": function (td, cellData, rowData, row, col) {
                        $(td).addClass('img');
                    },
                    "orderable": false,
                },
                // Type
                // {
                //     "targets": 1,
                //     "render": function (data, type, row) {
                //         return row[4];
                //     },
                //     "createdCell": function (td, cellData, rowData, row, col) {
                //         $(td).addClass('d-none d-lg-table-cell');
                //     },
                //     "orderable": false,
                // },
                // Players
                {
                    "targets": 2,
                    "render": function (data, type, row) {
                        return row[7].toLocaleString();
                    },
                    "orderSequence": ["desc"],
                },
                // Followers
                {
                    "targets": 3,
                    "render": function (data, type, row) {
                        return row[10].toLocaleString();
                    },
                    "orderSequence": ["desc"],
                },
                // Score
                {
                    "targets": 4,
                    "render": function (data, type, row) {
                        return row[5] + '%';
                    },
                    "orderSequence": ["desc"],
                },
                // Price
                {
                    "targets": 5,
                    "render": function (data, type, row) {
                        return row[6];
                    },
                    "orderSequence": ["desc"],
                },

                // Link
                {
                    "targets": 6,
                    "render": function (data, type, row) {
                        if (row[8]) {
                            return '<a href="' + row[8] + '" target="_blank" rel="noopener"><i class="fas fa-link"></i></a>';
                        }
                        return '';
                    },
                    "orderable": false,
                },
            ]
        };

        // Default form inputs
        const params = new URL(window.location).searchParams;

        const $platforms = $('#platforms');
        // if (params.getAll($platforms.attr('name')).length === 0) {
        //     setUrlParam($platforms.attr('name'), getOS());
        //     $platforms.trigger("chosen:updated");
        // }

        // Init table
        const searchFields = [
            $('#tags'),
            $('#genres'),
            $('#categories'),
            $('#developers'),
            $('#publishers'),
            $('#types'),
            $('#search'),
            $platforms,
            $priceElement,
            $scoreElement,
        ];

        $('table.table').gdbTable({
            tableOptions: options,
            searchFields: searchFields
        });

    })(jQuery, document);
}
