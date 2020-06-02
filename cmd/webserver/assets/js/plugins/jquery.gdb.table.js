;(function ($, window, document, user, undefined) {

    "use strict";

    // Create the defaults once
    const pluginName = "gdbTable";
    const defaults = {
        cache: true,
        searchFields: [],
        tableOptions: {
            "autoWidth": false,
            "dom": '<"dt-header"<"dt-pagination"p><"dt-info"i>>t<"dt-footer"<"dt-pagination"p>>r',
            "fixedHeader": true,
            "info": true, // Counts
            "processing": false,
            "language": {
                "processing": '<i class="fas fa-spinner fa-spin fa-3x fa-fw"></i>',
                "paginate": {
                    "next": '<i class="fas fa-chevron-right"></i>',
                    "previous": '<i class="fas fa-chevron-left"></i>',
                },
                "info": "_TOTAL_ rows (_PAGES_ pages)",
                "infoFiltered": " of _MAX_",
            },
            "lengthChange": false,
            "ordering": true,
            "pageLength": 100,
            "paging": true,
            "pagingType": 'gamedb',
            "searching": true,
            "stateSave": false,
            "search": {
                "smart": true,
            },
        },
    };

    // The actual plugin constructor
    function Plugin(element, options) {

        const $element = $(element);

        options = $.extend(true, {}, defaults, options);

        // Remove info text unless we want it
        if (!$element.hasClass('table-counts')) {
            options.tableOptions.dom = options.tableOptions.dom.replace("<\"dt-info\"i>", "");
        }

        // Add helper
        options.isAjax = function () {
            return $element.attr('data-path') != null;
        };

        const rowType = $element.attr('data-row-type');
        if (rowType) {
            options.tableOptions.language.info = "_TOTAL_ " + rowType + " (_PAGES_ pages)";
        }

        let initialValues = {};
        let currentValues = {};

        //
        if (options.isAjax()) {

            const ajax = function (data, callback, settings) {

                delete data.columns;
                data.search = currentValues;

                $.ajax({
                    data: data,
                    dataType: 'json',
                    cache: options.cache,
                    url: function () {
                        return $element.attr('data-path');
                    }(),
                    success: callback,
                    error: function (jqXHR, textStatus, errorThrown) {

                        data = {
                            "draw": "1",
                            "recordsTotal": "0",
                            "recordsFiltered": "0",
                            "data": [],
                            "limited": false
                        };

                        callback(data, textStatus, null);
                    },
                });
            };

            options = $.extend(true, {}, options, {
                tableOptions: {
                    processing: false,
                    serverSide: true,
                    orderMulti: false,
                    "ajax": ajax,
                }
            });

        } else {

            const disabled = $element.find('thead tr th[data-disabled]').map(function () {
                return $(this).index();
            }).get();

            options = $.extend(true, {}, options, {
                tableOptions: {
                    columnDefs: [
                        {
                            "orderable": false,
                            targets: disabled,
                        }
                    ],
                }
            });

            const limit = $element.attr('data-limit');
            if (limit) {
                options.tableOptions.pageLength = parseInt(limit);
            }
        }

        // Update initialValues from search field
        for (const $field of options.searchFields) {

            const name = getFieldName($field);
            const value = getFieldValue($field);

            if (name && value && value.length > 0) {
                initialValues[name] = value;
                currentValues[name] = value;
            }
        }

        // Update currentValues with url values
        const urlParams = new URL(window.location).searchParams;
        for (const $field of options.searchFields) {

            const name = getFieldName($field);

            if (name && urlParams.has(name)) {

                let value;
                if (isFieldMultiple($field)) {
                    value = urlParams.getAll(name);
                } else {
                    value = urlParams.get(name);
                }

                setFieldValue($field, value);
                currentValues[name] = value;

                if ($field.attr('data-hightlight')) {
                    this.highlight = value;
                }

                const colToSortBy = $field.attr('data-col-sort');
                if (colToSortBy && value) {
                    options.tableOptions.order = [[parseInt(colToSortBy), 'desc']];
                }
            }
        }

        // Add pagination url params to options
        // Commented out as the initial sort/order value becomes whatever is in the url
        // if (urlParams.has('p')) {
        //     const page = urlParams.get('p');
        //     options.tableOptions.displayStart = (page - 1) * options.tableOptions.pageLength;
        //     currentValues['p'] = page;
        // }
        // if (urlParams.has('s') && urlParams.has('o')) {
        //     const sort = urlParams.get('s');
        //     const order = urlParams.get('o');
        //     options.tableOptions.order = [[parseInt(sort), order]];
        //     currentValues['s'] = sort;
        //     currentValues['o'] = order;
        // }

        //
        this.options = options;
        this.element = element;
        this.user = user;
        this.initialValues = initialValues;
        this.currentValues = currentValues;
        this.urlParams = urlParams;
        this.init();
    }

    $.extend(Plugin.prototype, {
        init: function () {

            const parent = this;

            // Before AJAX
            $(this.element).on('preXhr.dt', function (e, settings, data) {

                $(parent.element).stop(true, true).fadeTo(200, 0.3); // Fade
            });

            // After AJAX
            $(this.element).on('xhr.dt', function (e, settings, json, xhr) {

                $(parent.element).stop(true, true).fadeTo(200, 1); // Unfade

                // Add donate button
                parent.limited = json.limited;
            });

            // Init table
            const dt = $(this.element).DataTable(this.options.tableOptions);
            this.dt = dt;
            this.initialValues.s = $.extend(true, [], dt.order()); // Using extend to copy, not reference

            // On Draw
            $(this.element).on('draw.dt', function (e, settings) {

                // Add donate button
                if (parent.limited) {
                    const bold = $('li.paginate_button.page-item.next.disabled').length > 0 ? 'font-weight-bold' : '';
                    const donate = parent.limited === 2
                        ? $('<li class="donate"><small><a href="/login"><i class="fas fa-unlock"></i> <span class="' + bold + '">Login for more!</span></a></small></li>')
                        : $('<li class="donate"><small><a href="/donate"><i class="fas fa-heart text-danger"></i> <span class="' + bold + '">Donate for more!</span></a></small></li>');
                    $(parent.element).parent().find('.dt-pagination ul.pagination').append(donate);
                }

                // Hide empty pagination
                const $paginationHeader = $(parent.element).parent().find('.dt-header .dt-pagination');
                const $paginationFooter = $(parent.element).parent().find('.dt-footer .dt-pagination');
                if (dt.page.info().pages <= 1) {
                    $paginationFooter.hide();
                    if (!$(parent.element).hasClass('table-counts')) {
                        $paginationHeader.hide();
                    }
                } else {
                    $paginationFooter.show();
                    $paginationHeader.show();
                }

                // Update URL
                // if ($(parent.element).is(":visible:not([data-ordering=false])")) {
                //
                //     const order = dt.order();
                //     if (order && order[0]) {
                //         setUrlParam('o', order[0][1]);
                //         setUrlParam('s', order[0][0]);
                //     }
                // }

                // Highlight
                if ($.fn.mark) {

                    const $markables = $('.markable');

                    $markables.unmark();

                    if (parent.highlight) {
                        $markables.mark(parent.highlight);
                    }
                }

                // Bold rows
                highLightOwnedGames($(parent.element));

                // Lazy load images
                observeLazyImages($(parent.element).find('img[data-lazy]'));

                // Fix broken images
                fixBrokenImages();
            });

            // On page change
            // $(this.element).on('page.dt', function (e, settings, processing) {
            //
            //     // Scroll on pagination click
            //     let padding = 15;
            //     if ($('.fixedHeader-floating').length > 0) {
            //         padding = padding + 48;
            //     }
            //     $('html, body').animate({
            //         scrollTop: $(this).prev().offset().top - padding
            //     }, 200);
            // });

            // On tab change
            $('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {

                // Fixes hidden fixed header tables
                $.each(window.gdbTables, function (index, value) {
                    value.fixedHeader.adjust();
                });

                //
                clearUrlParams();
            });

            // On search field change
            if (this.options.isAjax()) {
                for (const $field of this.options.searchFields) {

                    if ($field.hasClass('noUi-target')) { // Sliders

                        const name = $field.attr('data-name');
                        const slider = $field[0].noUiSlider;

                        slider.on('set', function (e) {

                            const value = slider.get();

                            if (name) {

                                parent.currentValues[name] = value;

                                setUrlParam(name, value);
                                // if (JSON.stringify(parent.initialValues[name]) === JSON.stringify(value)) {
                                //     deleteUrlParam(name);
                                // } else {
                                //     setUrlParam(name, value);
                                // }
                            }

                            if (typeof window.updateLabels == 'function') {
                                window.updateLabels();
                            }

                            dt.draw();
                        });


                        slider.on('update', function (e) {
                            if (typeof window.updateLabels == 'function') {
                                window.updateLabels();
                            }
                        });

                    } else { // Inputs

                        const name = $field.attr('name');

                        $field.on('change search', function (e) {

                            const colToSortBy = $field.attr('data-col-sort');
                            if (colToSortBy) {
                                dt.order([parseInt(colToSortBy), 'desc']);
                            }

                            let value;
                            if ($field.attr('type') === 'radio') {
                                value = $field.filter(':checked').val();
                            } else {
                                value = $field.val();
                            }

                            if ($field.attr('data-hightlight')) {
                                parent.highlight = value;
                            }

                            if (name) {

                                parent.currentValues[name] = value;

                                setUrlParam(name, value);
                                // if (JSON.stringify(parent.initialValues[name]) === JSON.stringify(value)) {
                                //     deleteUrlParam(name);
                                // } else {
                                //     setUrlParam(name, value);
                                // }
                            }

                            dt.draw();

                            return false;
                        });
                    }
                }
            } else {
                for (const $field of this.options.searchFields) {

                    const name = $field.attr('name');

                    $field.on('keyup search', function (e) {

                        const colToSortBy = $field.attr('data-col-sort');
                        if (colToSortBy) {
                            dt.order([parseInt(colToSortBy), 'desc']);
                        }

                        let value;
                        if ($field.attr('type') === 'radio') {
                            value = $field.filter(':checked').val();
                        } else {
                            value = $field.val();
                        }

                        if ($field.attr('data-hightlight')) {
                            parent.highlight = value;
                        }

                        if (name) {

                            parent.currentValues[name] = value;

                            setUrlParam(name, value);
                            // if (JSON.stringify(parent.initialValues[name]) === JSON.stringify(value)) {
                            //     deleteUrlParam(name);
                            // } else {
                            //     setUrlParam(name, value);
                            // }
                        }

                        dt.search($(this).val());
                        dt.draw();
                    });
                }
            }

            // Fixes scrolling to pagination on every click
            $('div.dt-pagination').on('focus', '.paginate_button > a', function () {
                $(this).trigger('blur');
            });

            // Local tables finish initializing before event handlers are attached,
            // so we trigger them again here.
            if (!this.options.isAjax()) {
                $(parent.element).trigger('draw.dt');
            }

            // Keep track of tables, so we can recalculate fixed headers on tab changes etc
            window.gdbTables = window.gdbTables || [];
            window.gdbTables.push();
        },
    });

    function getFieldValue($field) {

        if ($field.hasClass('noUi-target')) {

            return $field[0].noUiSlider.get();

        } else if ($field.attr('type') === 'radio') {

            return $field.filter(':checked').val();

        } else {

            return $field.val();
        }
    }

    function setFieldValue($field, value) {

        if ($field.hasClass('noUi-target')) {

            $field[0].noUiSlider.set(value);

        } else if ($field.attr('type') === 'radio') {

            $field.filter('[value' + value + ']').prop('checked', true);

        } else {

            $field.val(value);

            if ($field.hasClass('form-control-chosen')) {
                $field.trigger("chosen:updated");
            }
        }
    }

    function getFieldName($field) {

        if ($field.hasClass('noUi-target')) {

            return $field.attr('data-name');

        } else {

            return $field.attr('name');
        }
    }

    function isFieldMultiple($field) {

        if ($field.hasClass('noUi-target')) {

            return getFieldValue($field).length > 1;

        } else if ($field.prop('multiple')) {

            return true;

        } else {
            return false;
        }
    }

    $.fn[pluginName] = function (options) {
        return new Plugin(this, options).dt;
    };

    // Init local tables
    $('table.table.table-datatable').each(function (index) {
        $(this).gdbTable();
    });


})(jQuery, window, document, user);
